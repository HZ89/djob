package djob

import pb "version.uuzu.com/zhuhuipeng/djob/message"

// implementation of DjobServer interface

// JobInfo get the job from backend
func (a *Agent) JobInfo(name, region string) (*pb.Job, error) {
	return nil, nil
}

// ExecDone handle job exec return
func (a *Agent) ExecDone(execution *pb.Execution) error {
	return nil
}

func (a *Agent) ExecutionInfo(executionName string) (*pb.Execution, error) {
	return nil, nil
}
