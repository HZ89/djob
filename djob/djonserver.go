package djob

import (
	"github.com/Sirupsen/logrus"
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
	return nil
}
