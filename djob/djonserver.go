package djob

import pb "version.uuzu.com/zhuhuipeng/djob/message"

// implementation of DjobServer interface

// JobInfo func in the apibackend.go

// ExecDone handle job exec return
func (a *Agent) ExecDone(execution *pb.Execution) error {
	return nil
}
