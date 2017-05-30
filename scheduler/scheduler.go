package scheduler

import (
	"time"
)


type analyzer interface {
	Next() time.Time
}

type Job interface {
	IsDisposable() bool
}

type entry struct{
	Job Job
	Next time.Time
	Perv time.Time
	Analyzer analyzer
}

type Scheduler struct {
	NewJobCh <-chan *Job
	RunJobCh chan<- *Job
	StopCh <-chan struct{}

	addEntry chan *entry
	running bool
}
