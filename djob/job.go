package djob

import (
	"sync"
	"time"
)

type Job struct {
	// Job name. Must be unique
	Name string `json:"name"`
	// Job region. Job run in this region normally
	Region string `json:"region"`
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

	//Tags of the job, used for select node
	// Tags map[string]string `json:"tags"`

	// A expression used for filter agent node, job will run in the node when this expression is true
	Expression string `json:"expression"`
	//running status lock
	running sync.Mutex
	//Number of times to retry a job that failed an execution
	Retries uint `json:"retries"`

	//Other jobs that depend on this job.
	BeingDependentJobs []string `json:"being_dependent_jobs"`
	//This job depends on the job.
	ParentJob string `json:"parent_job"`

	//Parallel policy for this job
	Parallel bool `json:"parallel"`

	//Failover policy, can be "retry", "next_machine", "nil"
	Failover string `json:"failover"`

	// Last time running in this agent
	LastHandleAgent string `json:"last_handle_agent"`

	// disposable or not
	Disposable bool `json:"disposable"`

	// nodeName, this job be handled by the node
	NodeName string `json:"node_name"`

}

func (j *Job)IsDisposable() bool {
	return j.Disposable
}

type Execution struct {
	Name       string
	Cmd        string
	Output     []byte
	Succeed    bool
	StartTime  time.Time
	FinishTime time.Time
}
