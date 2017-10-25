package djob

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv/store"

	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

var APITimeOut = 30 * time.Second

func (a *Agent) JobModify(job *pb.Job) (*pb.Job, error) {

	if err := a.memStore.SetJob(job); err != nil {
		return nil, err
	}

	ej, err := a.store.GetJob(job.Name, job.Region)
	if err != nil && err != store.ErrKeyNotFound {
		return nil, err
	}
	var node string
	if ej != nil {
		node = ej.SchedulerNodeName
	} else {
		node, err = a.minimalLoadServer(job.Region)
		if err != nil {
			return nil, err
		}
	}
	resCh := make(chan *pb.QueryResult)
	errCh := make(chan error)
	go func() {
		result, err := a.sendNewJobQuery(job.Name, job.Region, node)
		if err != nil {
			errCh <- err
		}
		resCh <- result
	}()
	defer a.memStore.DeleteJob(job.Name, job.Region)

	select {
	case err := <-errCh:
		return nil, err
	case result := <-resCh:
		if result.Status != 0 {
			return nil, errors.New(result.Message)
		}
		rj, err := a.store.GetJob(job.Name, job.Region)
		if err != nil {
			return nil, err
		}
		return rj, nil
	case <-time.After(APITimeOut):
		return nil, ErrTimeOut
	}
}

func (a *Agent) JobInfo(name, region string) (job *pb.Job, err error) {
	job, err = a.store.GetJob(name, region)
	if err != nil && err != store.ErrKeyNotFound {
		log.Loger.WithError(err).Error("Agent: Get job from kvstore failed")
		return nil, err
	}
	if job == nil {
		job, err = a.memStore.GetJob(name, region)
		if err != nil {
			log.Loger.WithError(err).Error("Agent Get job from memstore failed")
			return nil, err
		}
	}
	return job, nil
}

// get all jobs which can be delete
// watch all jobs key
// send serf query to delete job
// wait timeout or delete finished
func (a *Agent) JobDelete(filter *pb.Job) (jobs []*pb.Job, err error) {
	if filter.Name == "" && filter.Region == "" {
		return nil, ErrArgs
	}
	jobs, err = a.JobList(filter)
	if err != nil {
		log.Loger.WithError(err).Error("Agent: Get job from kvstore failed")
		return nil, err
	}

	resCh := make(chan struct{})
	errCh := make(chan error)
	// send JobDeleteQuery
	for _, job := range jobs {
		go func() {
			a.sendJobDeleteQuery(job.Name, job.Region, job.SchedulerNodeName)
		}()

	}

	select {
	case err := <-errCh:
		return nil, err
	case result := <-resCh:
		if result.Status != 0 {
			return nil, errors.New(result.Message)
		}
		return job, nil
	case <-time.After(APITimeOut):
		return nil, ErrTimeOut
	}
}

func (a *Agent) JobList(filter *pb.Job) (jobs []*pb.Job, err error) {
	var regions []string
	if filter.Region == "" {
		regions, err = a.store.GetRegionList()
		if err != nil {
			return nil, err
		}
	} else {
		regions = append(regions, filter.Region)
	}
	if filter.Name == "" {
		for _, region := range regions {
			res, err := a.store.GetRegionJobs(region)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, res...)
		}
	} else {
		for _, region := range regions {
			res, err := a.store.GetJob(filter.Name, region)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, res)
		}

	}
	return
}

func (a *Agent) JobRun(name, region string) (*pb.Execution, error) {
	job, err := a.store.GetJob(name, region)
	if err != nil {
		return nil, err
	}
	ex := pb.Execution{
		SchedulerNodeName: job.SchedulerNodeName,
		Name:              job.Name,
		Region:            job.Region,
		Group:             time.Now().UnixNano(),
		Retries:           0,
	}
	log.Loger.WithFields(logrus.Fields{
		"name":   ex.Name,
		"region": ex.Region,
		"group":  ex.Group}).Debug("API: Prepare to run job")
	go a.sendRunJobQuery(&ex)
	return &ex, nil
}

func (a *Agent) JobStatus(name, region string) (*pb.JobStatus, error) {
	js, err := a.store.GetJobStatus(name, region)
	if err != nil {
		return nil, err
	}
	return js, nil
}
