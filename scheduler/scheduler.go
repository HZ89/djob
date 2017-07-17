package scheduler

import (
	"time"
	pb "local/djob/message"
	"errors"
	"sort"
)

var nameToIndex map[string]int

const DELETEJOB string = "DELETE"
const UPDATEJOB string = "UPDATE"

type analyzer interface {
	Next(time.Time) time.Time
}

type JobEvent struct {
	event string
	job   *pb.Job
}

type entry struct {
	Job      *pb.Job
	Next     time.Time
	Perv     time.Time
	Analyzer analyzer
}

type Scheduler struct {
	JobEventCh <-chan *JobEvent
	RunJobCh   chan *pb.Execution
	stopCh     chan struct{}

	running bool
	entries []*entry
}

type byTime []*entry

func (b byTime) Len() int      { return len(b) }
func (b byTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byTime) Less(i, j int) bool {
	if b[i].Next.IsZero() {
		return false
	}
	if b[j].Next.IsZero() {
		return false
	}
	return b[i].Next.Before(b[j].Next)
}

func deleteEntryByName(b byTime, name string) (byTime, error) {
	i, exists := nameToIndex[name]
	if exists {
		if i < b.Len() && b[i].Job.Name == name {
			// delete i
			copy(b[i:], b[i+1:])
			b[b.Len()-1] = nil
			b = b[:b.Len()-1]
		}
		delete(nameToIndex, name)
		return b, nil
	}
	return nil, errors.New("Con't find job")
}

func newEntery(job *pb.Job) *entry {
	analyzer, _ := prepare(job.Schedule)
	return &entry{
		Job:      job,
		Analyzer: analyzer,
	}
}

func New(jobeventCh <-chan *JobEvent, runjobCh chan *pb.Execution) *Scheduler{
	return &Scheduler{
		JobEventCh: jobeventCh,
		RunJobCh: runjobCh,
		stopCh: make(chan struct{}),
		entries: nil,
		running: false,
	}
}

func (s *Scheduler) send(job *pb.Job) {

}

func (s *Scheduler) run() {
	now := time.Now().Local()
	for _, e := range s.entries {
		e.Next = e.Analyzer.Next(now)
	}

	for {
		sort.Sort(byTime(s.entries))
		for i, e := range s.entries {
			nameToIndex[e.Job.Name] = i
		}
		var movement time.Time
		if len(s.entries) == 0 || s.entries[0].Next.IsZero() {
			movement = now.AddDate(10, 0, 0)
		} else {
			movement = s.entries[0].Next
		}

		select {
		case now = <-time.After(movement.Sub(now)):
			for _, e := range s.entries {
				if e.Next != movement {
					break
				}
				go s.send(e.Job)
				e.Perv = e.Next
				e.Next = e.Analyzer.Next(movement)
			}
			continue
		case jobevent := <-s.JobEventCh:
			if jobevent.event == UPDATEJOB {
				newEntery := newEntery(jobevent.job)
				// delete the job if it exits
				s.entries, _ = deleteEntryByName(s.entries, jobevent.job.Name)
				s.entries = append(s.entries, newEntery)
			}
			if jobevent.event == DELETEJOB {
				s.entries, _ = deleteEntryByName(s.entries, jobevent.job.Name)
			}
		case <-s.stopCh:
			return
		}
		now = time.Now().Local()
	}
}

func (s *Scheduler) Stop() {
	s.stopCh <- struct{}{}
	s.running = false
}

func (s *Scheduler) Start() {
	s.running = true
	s.run()
}
