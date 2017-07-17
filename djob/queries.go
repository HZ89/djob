package djob

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv/store"
	"github.com/hashicorp/serf/serf"
	pb "local/djob/message"
	"github.com/gogo/protobuf/proto"
)

const (
	QueryNewJob    = "job:new"
	QueryRunJob    = "job:run"
	QueryRPCConfig = "rpc:config"
)

//type NewJobQueryParams struct {
//	JobName string
//	Region string
//}
//
//type RunJobQueryParams struct {
//	JobName string
//	SchedulerNodeName string
//}

func (a *Agent) sendNewJobQuery(jobName string) {
	var params *serf.QueryParam
	job, err := a.store.GetJob(jobName)
	if err != nil {
		if err == store.ErrKeyNotFound {
			Log.WithFields(logrus.Fields{
				"jobName": jobName,
			}).Debug("job can not found")
		}
	}
	params, err = a.createSerfQueryParam(job.Expression)
	if err != nil {
		if err == ErrCanNotFoundNode {
			Log.WithFields(logrus.Fields{
				"jobName": job.Name,
				"jobRegion": job.Region,
				"jobExp": job.Expression,
			}).Debug(err)
		}
		Log.Warn(err)
		return
		//var regionFilter map[string]string
		//regionFilter["region"] = job.Region
		//params = &serf.QueryParam{
		//	FilterTags: regionFilter,
		//	RequestAck: true,
		//}
	}
	qp := &pb.NewJobQueryParams{
		Name: job.Name,
		Region: job.Region,
	}
	qpPb, _ := proto.Marshal(qp)

	Log.WithFields(logrus.Fields{
		"query_name": QueryNewJob,
		"job_name": job.Name,
		"job_region": job.Region,
		"playload": qp.String(),
	}).Debug("agent: Sending query")

	qr, err := a.serf.Query(QueryNewJob, qpPb, params)
	if err != nil {
		Log.WithField("query", QueryNewJob).WithError(err).Fatal("agent: Sending query error")
	}
	defer qr.Close()

	ackCh := qr.AckCh()
	respCh := qr.ResponseCh()
	for !qr.Finished() {
		select {
		case ack, ok := <-ackCh:
			if ok {
				Log.WithFields(logrus.Fields{
					"query": QueryRunJob,
					"from": ack,
				}).Debug("agent: Received ack")
			}

		case resp, ok:=<-respCh:
			if ok {
				Log.WithFields(logrus.Fields{
					"query": QueryRunJob,
					"payload": string(resp.Payload),
				}).Debug("agent: Received response")
			}

		}
	}
}

func (a *Agent) receiveNewJobQuery(query *serf.Query) {
	var params pb.NewJobQueryParams
	if err := proto.Unmarshal(query.Payload, params); err != nil {
		Log.WithFields(logrus.Fields{
			"query":   query.Name,
			"payload": string(query.Payload),
		}).WithError(err).Error("agent: Server add new job memberevent")
	}
	job, err := a.store.GetJob(params.JobName)
	if err != nil {
		Log.WithFields(logrus.Fields{
			"query":   query.Name,
			"payload": string(query.Payload),
		}).WithError(err).Error("agent: Server add new job memberevent")
	}
	// try to lock this job.
	locker, err := a.lockJob(job.Name)

	if err != nil {
		if err == ErrLockTimeout{
			Log.WithField("jobName:", job.Name).WithError(err).Debug("agent: try lock a job")
		}
		Log.WithFields(logrus.Fields{
			"query":   query.Name,
			"payload": string(query.Payload),
		}).WithError(err).Debug("agent: Server add new job memberevent")
		return
	}

	a.jobLockers[job.Name] = locker

	a.newJobCh <- job
	Log.Infof("agent: send job %s to newJobCh", job.Name)
	Log.WithFields(logrus.Fields{
		"query": query.Name,
		"payload": string(query.Payload),
	}).Debug("agent: send job to newJobCh")
}

func (a *Agent) sendRunJobQuery(job *pb.Job) {

}
