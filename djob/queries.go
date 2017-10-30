package djob

import (
	"fmt"
	"math/rand"

	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/serf/serf"
	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

const (
	QueryRunJob    = "job:run"
	QueryRPCConfig = "rpc:config"
	QueryJobCount  = "job:count"
)

func (a *Agent) sendJobCountQuery(region string) ([]*pb.JobCountResp, error) {

	params, err := a.createSerfQueryParam(fmt.Sprintf("server=='true'&&region=='%s'", region))
	if err != nil {
		return nil, err
	}
	params.FilterTags = map[string]string{
		"server": "true",
		"region": region,
	}
	qr, err := a.serf.Query(QueryJobCount, nil, params)
	if err != nil {
		log.Loger.WithFields(logrus.Fields{
			"query": QueryJobCount,
			"error": err,
		}).Fatal("Agent: Error sending serf query")
	}
	defer qr.Close()

	ackCh := qr.AckCh()
	respCh := qr.ResponseCh()
	var payloads [][]byte
	for {
		if len(payloads) == len(params.FilterNodes) {
			break
		}
		select {

		case ack, ok := <-ackCh:
			if ok {
				log.Loger.WithFields(logrus.Fields{
					"query": QueryJobCount,
					"from":  ack,
				}).Debug("Agent: Received ack")
			}
		case resp, ok := <-respCh:
			if ok {
				log.Loger.WithFields(logrus.Fields{
					"query": QueryJobCount,
					"resp":  string(resp.Payload),
				}).Debug("Agent: Received resp")
				payloads = append(payloads, resp.Payload)
			}
		}
	}

	var r []*pb.JobCountResp
	for _, v := range payloads {
		var p pb.JobCountResp
		proto.Unmarshal(v, &p)
		r = append(r, &p)
	}
	return r, nil
}

func (a *Agent) receiveJobCountQuery(query *serf.Query) {
	resp := pb.JobCountResp{
		Name:  a.config.Nodename,
		Count: int64(a.scheduler.JobCount()),
	}
	respPb, _ := proto.Marshal(&resp)
	log.Loger.WithFields(logrus.Fields{
		"name":   resp.Name,
		"count":  resp.Count,
		"respPb": string(respPb),
	}).Debug("Agent: respone of job:count")
	if err := query.Respond(respPb); err != nil {
		log.Loger.WithError(err).Fatal("Agent: serf quert Responed failed")
	}
	return
}

func (a *Agent) sendRunJobQuery(ex *pb.Execution) {

	job, err := a.store.GetJob(ex.Name, ex.Region)
	if err != nil {
		log.Loger.WithError(err).Fatal("Agent GetJob failed")
	}

	exPb, err := proto.Marshal(ex)
	if err != nil {
		log.Loger.WithError(err).Fatal("Agent: Encode failed")
	}

	var params *serf.QueryParam
	if ex.Retries < 1 {
		params, err = a.createSerfQueryParam(job.Expression)
	} else {
		params = &serf.QueryParam{
			FilterNodes: []string{ex.RunNodeName},
			RequestAck:  true,
		}
	}

	if err != nil {
		log.Loger.WithFields(logrus.Fields{
			"Query": QueryRunJob,
			"stage": "prepare",
		}).WithError(err).Error("Agent: Create serf query param failed")
		return
	}

	log.Loger.WithFields(logrus.Fields{
		"query_name": QueryRunJob,
		"job_name":   job.Name,
		"job_region": job.Region,
		"payload":    string(exPb),
	}).Debug("Agent: Sending query")

	qr, err := a.serf.Query(QueryRunJob, exPb, params)
	if err != nil {
		log.Loger.WithField("query", QueryRunJob).WithError(err).Fatal("Agent: Sending query error")
	}
	defer qr.Close()

	ackCh := qr.AckCh()
	respCh := qr.ResponseCh()
	for !qr.Finished() {
		select {
		case ack, ok := <-ackCh:
			if ok {
				log.Loger.WithFields(logrus.Fields{
					"query": QueryRunJob,
					"from":  ack,
				}).Debug("Agent: Received ack")
			}

		case resp, ok := <-respCh:
			if ok {
				log.Loger.WithFields(logrus.Fields{
					"query":   QueryRunJob,
					"payload": string(resp.Payload),
				}).Debug("Agent: Received response")
			}

		}
	}
	log.Loger.WithFields(logrus.Fields{
		"query": QueryRunJob,
	}).Debug("Agent: Done receiving acks and responses")

}

func (a *Agent) receiveRunJobQuery(query *serf.Query) {
	var ex pb.Execution
	if err := proto.Unmarshal(query.Payload, &ex); err != nil {
		log.Loger.WithError(err).Fatal("Agent: Error decode query payload")
	}

	log.Loger.WithFields(logrus.Fields{
		"job":    ex.Name,
		"region": ex.Region,
		"group":  ex.Group,
	}).Info("Agent: Starting job")

	ip, port, err := a.sendGetRPCConfigQuery(ex.SchedulerNodeName)
	if err != nil {
		log.Loger.WithError(err).Fatal("Agent: get rpc config failed")
	}

	rpcc := a.newRPCClient(ip, port)
	defer rpcc.Shutdown()

	job, err := rpcc.GetJob(ex.Name, ex.Region)
	if err != nil {
		log.Loger.WithError(err).Fatal("Agent: rpc call GetJob Failed")
	}
	logrus.WithFields(logrus.Fields{
		"job":    job.Name,
		"region": job.Region,
		"cmd":    job.Command,
	}).Debug("Agent: Got job by rpc call GetJob")

	go func() {
		if err = a.execJob(job, &ex); err != nil {
			log.Loger.WithError(err).Error("Proc: Exec job Failed")
		}
		log.Loger.WithFields(logrus.Fields{
			"job":    ex.Name,
			"region": ex.Region,
			"group":  ex.Group,
		}).Info("Agent: Job done")
	}()
}

func (a *Agent) sendGetRPCConfigQuery(nodeName string) (string, int, error) {
	params := &serf.QueryParam{
		FilterNodes: []string{nodeName},
	}

	qr, err := a.serf.Query(QueryRPCConfig, nil, params)
	if err != nil {
		log.Loger.WithFields(logrus.Fields{
			"query": QueryRPCConfig,
			"error": err,
		}).Fatal("Agent: Error sending serf query")
	}

	defer qr.Close()

	ackCh := qr.AckCh()
	respCh := qr.ResponseCh()
	var payloads [][]byte
	for {
		if len(payloads) == len(params.FilterNodes) {
			break
		}
		select {
		case ack, ok := <-ackCh:
			if ok {
				log.Loger.WithFields(logrus.Fields{
					"query": QueryRPCConfig,
					"from":  ack,
				}).Debug("Agent: Received ack")
			}
		case resp, ok := <-respCh:
			if ok {
				payloads = append(payloads, resp.Payload)
				log.Loger.WithFields(logrus.Fields{
					"query":   QueryRPCConfig,
					"from":    resp.From,
					"payload": string(resp.Payload),
				}).Debug("Agent: Received response")
			}
		}
	}
	log.Loger.WithFields(logrus.Fields{
		"query": QueryRPCConfig,
	}).Debug("Agent: Received ack and response Done")

	var payload pb.GetRPCConfigResp
	if err := proto.Unmarshal(payloads[rand.Intn(len(payloads))], &payload); err != nil {
		log.Loger.WithError(err).Error("Agent: Payload decode failed")
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
		log.Loger.WithError(err).Fatal("Agent: serf query Respond error")
	}
	return
}
