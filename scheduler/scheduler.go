package scheduler

import (
	"errors"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"sort"
	"time"
)


type analyzer interface {
	Next(time.Time) time.Time
}

type jobEvent struct {
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
	addEntry    chan *entry
	deleteEntry chan string
	RunJobCh    chan *pb.Job
	stopCh      chan struct{}

	running     bool
	entries     []*entry
	nameToIndex map[string]int
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

func newEntery(job *pb.Job) (*entry, error) {
	analyzer, err := Prepare(job.Schedule)
	if err != nil {
		return nil, err
	}
	return &entry{
		Job:      job,
		Analyzer: analyzer,
	}, nil
}

func New(runjobCh chan *pb.Job) *Scheduler {
	return &Scheduler{
		addEntry:    make(chan *entry),
		deleteEntry: make(chan string),
		RunJobCh:    runjobCh,
		stopCh:      make(chan struct{}),
		entries:     nil,
		nameToIndex: make(map[string]int),
		running:     false,
	}
}

func (s *Scheduler) deleteEntryByName(name string) error {
	i, exists := s.nameToIndex[name]
	if exists {
		if i < len(s.entries) && s.entries[i].Job.Name == name {
			// delete i
			copy(s.entries[i:], s.entries[i+1:])
			s.entries[len(s.entries)-1] = nil
			s.entries = s.entries[:len(s.entries)-1]
		}
		delete(s.nameToIndex, name)
		return nil
	}
	return errors.New("Con't find job")
}

/*
TODO: 重写runJob，添加一个backend interface，包含异步任务执行方法，同步任务执行方法和任务执行结果查询方法
TODO: 添加 基于任务执行结果的调度
 */
func (s *Scheduler) runJob(job *pb.Job) {
	s.RunJobCh <- job
}

func (s *Scheduler) run() {
	now := time.Now().Local()
	for _, e := range s.entries {
		e.Next = e.Analyzer.Next(now)
	}

	for {
		sort.Sort(byTime(s.entries))
		for i, e := range s.entries {
			s.nameToIndex[e.Job.Name] = i
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
				go s.runJob(e.Job)
				e.Perv = e.Next
				e.Next = e.Analyzer.Next(movement)
				if e.Next.IsZero() {
					s.deleteEntryByName(e.Job.Name)
				}
			}
			continue
		case e := <-s.addEntry:
			s.deleteEntryByName(e.Job.Name)
			s.entries = append(s.entries, e)
		case name := <-s.deleteEntry:
			s.deleteEntryByName(name)
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
	go s.run()
}

func (s *Scheduler) AddJob(job *pb.Job) error {
	now := time.Now().Local()
	e, err := newEntery(job)
	if err != nil {
		return err
	}
	e.Next = e.Analyzer.Next(now)
	s.addEntry <- e
	return nil
}

func (s *Scheduler) DeleteJob(name string) {
	s.deleteEntry <- name
}
