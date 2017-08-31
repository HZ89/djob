package djob

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv/store"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

// implementation of DjobServer interface

// JobInfo func in the apibackend.go

// ExecDone handle job exec return
func (a *Agent) ExecDone(ex *pb.Execution) (err error) {
	Log.WithFields(logrus.Fields{
		"JobName":     ex.JobName,
		"Region":      ex.Region,
		"Group":       ex.Group,
		"RunNodeName": ex.RunNodeName,
	}).Debug("RPC: Save Execution to backend")

	if err = a.sqlStore.SetExec(ex); err != nil {
		Log.WithError(err).WithFields(logrus.Fields{
			"JobName":     ex.JobName,
			"Region":      ex.Region,
			"Group":       ex.Group,
			"RunNodeName": ex.RunNodeName,
		}).Error("RPC: Save Execution to backend failed")
		return
	}
	var l store.Locker
	l, err = a.lock(ex.JobName, ex.Region, &pb.JobStatus{})
	if err != nil {
		return
	}
	defer l.Unlock()
	if err = a.store.SetJobStatus(ex); err != nil {
		Log.WithError(err).WithFields(logrus.Fields{
			"JobName":     ex.JobName,
			"Region":      ex.Region,
			"Group":       ex.Group,
			"RunNodeName": ex.RunNodeName,
		}).Error("RPC: set JobStatus to kv store failed")
		return
	}

	return nil
}
