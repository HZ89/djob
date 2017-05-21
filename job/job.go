package job

import "time"

type Job struct {
	Name string
	SchedulerName string
}


type Execution struct {
	Name string
	Cmd string
	Output []byte
	Succeed bool
	StartTime time.Time
	FinishTime time.Time
}