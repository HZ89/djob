package djob

import (
	pb "version.uuzu.com/zhuhuipeng/djob/message"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/serf/serf"
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

// sendNreJobQuery func used to notice a server there is a now job need to be add
func (a *Agent) sendNewJobQuery(jobName, region, serverNodeName string) {

	params := &serf.QueryParam{
		FilterNodes: []string(serverNodeName),
		FilterTags: map[string]string{
			"server": "true",
			"region": region,
		},
		RequestAck: true,
	}

	qp := &pb.NewJobQueryParams{
		Name:           jobName,
		Region:         region,
		SourceNodeName: a.config.Nodename,
	}
	qpPb, _ := proto.Marshal(qp)

	Log.WithFields(logrus.Fields{
		"query_name": QueryNewJob,
		"job_name":   jobName,
		"job_region": region,
		"playload":   qp.String(),
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
					"from":  ack,
				}).Debug("agent: Received ack")
			}

		case resp, ok := <-respCh:
			if ok {
				Log.WithFields(logrus.Fields{
					"query":   QueryRunJob,
					"payload": string(resp.Payload),
				}).Debug("agent: Received response")
			}

		}
	}
}

func (a *Agent) receiveNewJobQuery(query *serf.Query) {
	var params *pb.NewJobQueryParams
	if err := proto.Unmarshal(query.Payload, params); err != nil {
		Log.WithFields(logrus.Fields{
			"query":   query.Name,
			"payload": string(query.Payload),
		}).WithError(err).Error("Agent: Server add new job memberevent")
	}
	if params.Region != a.config.Region {
		Log.WithFields(logrus.Fields{
			"name":   params.Name,
			"region": params.Region,
		}).Debug("Agent: receive a job from other region")
		return
	}
	ip, port, err := a.sendGetRPCConfigQuery(params.SourceNodeName)
	if err != nil {
		Log.WithFields(logrus.Fields{
			"SourceNodeName": params.SourceNodeName,
		}).WithError(err).Error("Agent: Get RPC config Failed")
		return
	}

	rpcClient := a.newRPCClient(ip, port)
	defer rpcClient.Shutdown()

	job, err := rpcClient.GetJob(params.Name)
	if err != nil {
		Log.WithFields(logrus.Fields{
			"Name":    query.Name,
			"RPCAddr": fmt.Sprintf("%s:%d", ip, port),
		}).WithError(err).Error("Agent: RPC Client get job failed")
	}

	a.mutex.Lock()
	if _, exist := a.jobLockers[job.Name]; !exist {
		// try to lock this job.
		locker, err := a.lockJob(job.Name, job.Region)
		if err != nil {
			Log.WithFields(logrus.Fields{
				"JobName": job.Name,
				"Region":  job.Region,
			}).WithError(err).Fatal("Agent: try lock a job")
		}

		a.jobLockers[job.Name] = locker
		job.SchedulerNodeName = a.config.Nodename
	}
	a.mutex.Unlock()

	if err := a.store.SetJob(job); err != nil {
		Log.WithFields(logrus.Fields{
			"JobName": job.Name,
			"Region":  job.Region,
		}).WithError(err).Fatal("Agent: Save job to store failed")
	}

	a.scheduler.DeleteJob(job.Name)

	// add job
	err = a.scheduler.AddJob(job)
	if err != nil {
		Log.WithFields(logrus.Fields{
			"job_name":  job.Name,
			"scheduler": job.Schedule,
		}).WithError(err).Fatal("Add Job to scheduler failed")
	}
	Log.Infof("agent: send job %s to newJobCh", job.Name)
	Log.WithFields(logrus.Fields{
		"query":   query.Name,
		"payload": string(query.Payload),
	}).Debug("agent: send job to newJobCh")
}

func (a *Agent) sendRunJobQuery(job *pb.Job) {

}

func (a *Agent) receiveRunJobQuery(query *serf.Query) {

}

func (a *Agent) sendGetRPCConfigQuery(nodeName string) (string, int, error) {
	params := &serf.QueryParam{
		FilterNodes: []string(nodeName),
		FilterTags:  map[string]string{"server": "true"},
		RequestAck:  true,
	}

	qr, err := a.serf.Query(QueryRPCConfig, nil, params)
	if err != nil {
		Log.WithFields(logrus.Fields{
			"query": QueryRPCConfig,
			"error": err,
		}).Fatal("Agent: Error sending serf query")
	}

	defer qr.Close()

	ackCh := qr.AckCh()
	respCh := qr.ResponseCh()
	var payloadPb []byte
	for !qr.Finished() {
		select {
		case ack, ok := <-ackCh:
			if ok {
				Log.WithFields(logrus.Fields{
					"query": QueryRPCConfig,
					"from":  ack,
				}).Debug("Agent: Received ack")
			}
		case resp, ok := <-respCh:
			if ok {
				Log.WithFields(logrus.Fields{
					"query":   QueryRPCConfig,
					"from":    resp.From,
					"payload": string(resp.Payload),
				}).Debug("Agent: Received response")
				payloadPb = resp.Payload
			}
		}
	}
	Log.WithFields(logrus.Fields{
		"query": QueryRPCConfig,
	}).Debug("Agent: Received ack and response Done")

	var payload *pb.GetRPCConfigResp
	if err := proto.Unmarshal(payloadPb, payload); err != nil {
		Log.WithError(err).Error("Agent: Payload decode failed")
		return "", 0, nil
	}
	return payload.Ip, int(payload.Port), nil
}

func (a *Agent) receiveGetRPCConfigQuery(query *serf.Query) {
	resp := &pb.GetRPCConfigResp{
		Ip:   a.config.RPCBindIP,
		Port: int32(a.config.RPCBindPort),
	}
	respPb, _ := proto.Marshal(resp)
	if err := query.Respond(respPb); err != nil {
		Log.WithError(err).Errorf("Agent: serf query Respond error")
	}
	return
}
