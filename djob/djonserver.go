package djob

import pb "local/djob/message"

// implementation of DjobServer interface

// JobInfo get the job from backend
func (a *Agent) JobInfo(jobName string) (*pb.Job, error) {
	return nil, nil
}

// ExecDone handle job exec return
func (a *Agent) ExecDone(execution *pb.Execution) error {
	return nil
}

func (a *Agent) ExecutionInfo(executionName string) (*pb.Execution, error) {
	return nil, nil
}
