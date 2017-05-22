package job

import (
	"time"
	"sync"
	"github.com/docker/libkv/store"
)

type Job struct {
	// Job name. Must be unique
	Name string `json:"name"`
	// Name of scheduler, used for select scheduler
	SchedulerName string `json:"scheduler_name"`
	// The parameter to scheduler
	ScheduleParameter map[string]string `json:"schedule_parameter"`
	// Use shell to run command
	Shell bool `json:"shell"`
	// Command to run
	Command string `json:"command"`
	// Number of successful executions of this job
	SuccessCount int `json:"success_count"`
	// Number of errors of this job
	ErrorCount int `json:"error_count"`
	// Last time this jon executed successful
	LastSuccess time.Time `json:"last_success"`
	//Last time this job failed
	LastError time.Time `json:"last_error"`
	//Tags of the job, used for select server
	Tags map[string]string `json:"tags"`
	//running status lock
	running sync.Mutex
	//Number of times to retry a job that failed an execution
	Retries uint `json:"retries"`
	//To prevent being edited repeatedly
	locker store.Locker

	//Other jobs that depend on this job.
	BeingDependentJobs []string `json:"being_dependent_jobs"`
	//This job depends on the job.
	ParentJob string `json:"parent_job"`

	//Parallel policy for this job
	Parallel bool `json:"parallel"`

	//Failover policy, can be "retry", "next_machine", "nil"
	Failover string `json:"failover"`
}


type Execution struct {
	Name string
	Cmd string
	Output []byte
	Succeed bool
	StartTime time.Time
	FinishTime time.Time
}