package djob

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Sirupsen/logrus"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/store"
)

// implementation of Operator interface

// SendBackExecution handle job exec return
func (a *Agent) SendBackExecution(ex *pb.Execution) (err error) {
	log.Loger.WithFields(logrus.Fields{
		"JobName":     ex.Name,
		"Region":      ex.Region,
		"Group":       ex.Group,
		"RunNodeName": ex.RunNodeName,
	}).Debug("RPC: Save Execution to backend")

	_, _, err = a.operationMiddleLayer(ex, pb.Ops_ADD, nil)
	if err != nil {
		log.Loger.WithError(err).WithFields(logrus.Fields{
			"JobName":     ex.Name,
			"Region":      ex.Region,
			"Group":       ex.Group,
			"RunNodeName": ex.RunNodeName,
		}).Error("RPC: Save Execution to backend failed")
		return
	}
	status := &pb.JobStatus{
		Name:   ex.Name,
		Region: ex.Region,
	}

	statusLocker, err := a.store.Lock(status, store.W, a.config.Nodename)
	if err != nil {
		return err
	}
	defer statusLocker.Unlock()

	out, _, err := a.operationMiddleLayer(status, pb.Ops_READ, nil)
	if err != nil && err != errors.ErrNotFound {
		return err
	}

	if len(out) != 0 {
		es, ok := out[0].(*pb.JobStatus)
		if !ok {
			log.Loger.Fatal(fmt.Sprintf("RPC: SendBackExecution want a JobStatus, but %v", reflect.TypeOf(out[0])))
		}

		status.LastError = es.LastError
		status.LastSuccess = es.LastSuccess
		status.SuccessCount = es.SuccessCount
		status.ErrorCount = es.ErrorCount
	}

	status.LastHandleAgent = ex.RunNodeName

	if ex.Succeed {
		status.SuccessCount += 1
		status.LastSuccess = time.Unix(0, ex.FinishTime).String()
	} else {
		status.ErrorCount += 1
		status.LastError = time.Unix(0, ex.FinishTime).String()
	}

	_, _, err = a.operationMiddleLayer(status, pb.Ops_MODIFY, nil)
	if err != nil {
		log.Loger.WithError(err).WithFields(logrus.Fields{
			"JobName":     status.Name,
			"Region":      status.Region,
			"Group":       ex.Group,
			"RunNodeName": ex.RunNodeName,
		}).Error("RPC: set JobStatus to kv store failed")
		return
	}

	return nil
}

// get Job object
func (a *Agent) GetJob(name, region string) (*pb.Job, error) {
	out, _, err := a.operationMiddleLayer(&pb.Job{Name: name, Region: region}, pb.Ops_READ, nil)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		log.Loger.WithFields(logrus.Fields{
			"Name":   name,
			"Region": region,
		}).Warn("RPC: GetJob the return nothing")
		return nil, errors.ErrNotExist
	}
	if len(out) != 1 {
		log.Loger.WithFields(logrus.Fields{
			"Name":   name,
			"Region": region,
		}).Warn("RPC: GetJob return job object is not unique")
	}
	if t, ok := out[0].(*pb.Job); ok {
		return t, nil
	}
	log.Loger.WithFields(logrus.Fields{
		"Name":   name,
		"Region": region,
	}).Fatalf("RPC: GetJob want a %v, but get %v", reflect.TypeOf(&pb.Job{}), reflect.TypeOf(out[0]))
	return nil, errors.ErrType
}

// forwarding ops to remote or perform it in local
func (a *Agent) PerformOps(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int, error) {
	return a.operationMiddleLayer(obj, ops, search)
}
