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

	if err = a.setExec(ex); err != nil {
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

func (a *Agent) getExecs(jobName, region, node string, group int64) (exs []*pb.Execution, err error) {
	q := map[string]interface{}{
		"job_name": jobName,
		"region":   region,
	}
	if group > 0 {
		q["group"] = group
	}
	if node != "" {
		q["run_node_name"] = node
	}
	if err = a.db.Where(q).Find(&exs).Error; err != nil {
		return nil, err
	}
	return
}

func (a *Agent) getExec(jobName, region, node string, group int64) (ex *pb.Execution, err error) {
	var exs []*pb.Execution
	exs, err = a.getExecs(jobName, region, node, group)
	if err != nil {
		return
	}
	if len(exs) > 0 {
		return exs[0], nil
	}
	return
}

func (a *Agent) setExec(ex *pb.Execution) error {
	ej, err := a.getExec(ex.JobName, ex.Region, ex.RunNodeName, ex.Group)
	if err != nil {
		return err
	}
	if ej == nil {
		if err := a.db.Create(ex).Error; err != nil {
			return err
		}
	} else {
		if err := a.db.Save(ex).Error; err != nil {
			return err
		}
	}
	return nil
}
