package djob

import (
	"github.com/Sirupsen/logrus"
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

	if err = a.sqlStore.SetExec(ex); err != nil {
		log.Loger.WithError(err).WithFields(logrus.Fields{
			"JobName":     ex.Name,
			"Region":      ex.Region,
			"Group":       ex.Group,
			"RunNodeName": ex.RunNodeName,
		}).Error("RPC: Save Execution to backend failed")
		return
	}

	l, err := a.store.Lock(ex, store.W, "")
	if err != nil {
		return
	}
	defer l.Unlock()
	if err = a.store.SetJobStatus(ex); err != nil {
		log.Loger.WithError(err).WithFields(logrus.Fields{
			"JobName":     ex.Name,
			"Region":      ex.Region,
			"Group":       ex.Group,
			"RunNodeName": ex.RunNodeName,
		}).Error("RPC: set JobStatus to kv store failed")
		return
	}

	return nil
}
