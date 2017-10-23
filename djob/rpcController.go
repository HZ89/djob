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

// implementation of DjobServer interface
// JobInfo func in the apibackend.go
// ExecDone handle job exec return
func (a *Agent) ExecDone(ex *pb.Execution) (err error) {
	log.Loger.WithFields(logrus.Fields{
		"JobName":     ex.Name,
		"Region":      ex.Region,
		"Group":       ex.Group,
		"RunNodeName": ex.RunNodeName,
	}).Debug("RPC: Save Execution to backend")

	_, _, err = a.DoOps(ex, pb.Ops_ADD, nil)
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
			log.Loger.Fatal(fmt.Sprintf("RPC: ExecDone want a JobStatus, but %v", reflect.TypeOf(out[0])))
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

func (a *Agent) DoOps(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int, error) {
	return a.operationMiddleLayer(obj, ops, search)
}
