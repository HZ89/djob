package djob

import (
	"errors"
	"github.com/docker/libkv/store"
	"time"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

var APITimeOut = 5 * time.Second

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
	resCh := make(chan *pb.Result)
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
		Log.WithError(err).Error("Agent: Get job from kvstore failed")
		return nil, err
	}
	if job == nil {
		job, err = a.memStore.GetJob(name, region)
		if err != nil {
			Log.WithError(err).Error("Agent Get job from memstore failed")
			return nil, err
		}
	}
	return job, nil
}

func (a *Agent) JobDelete(name, region string) (job *pb.Job, err error) {
	job, err = a.store.GetJob(name, region)
	if err != nil {
		Log.WithError(err).Error("Agent: Get job from kvstore failed")
		return nil, err
	}

	resCh := make(chan *pb.Result)
	errCh := make(chan error)

	go func() {
		resp, err := a.sendJobDeleteQuery(name, region, job.SchedulerNodeName)
		if err != nil {
			errCh <- err
		}
		resCh <- resp
	}()

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

func (a *Agent) JobList(region string) ([]*pb.Job, error) {
	jobs, err := a.store.GetJobList(region)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
