package djob

import (
	"errors"
	"github.com/docker/libkv/store"
	"time"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

var APITimeOut = 5 * time.Second

func (a *Agent) JobModify(job *pb.Job) (*pb.RespJob, error) {

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
		node = a.minimalLoadServer()
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

	select {
	case err := <-errCh:
		return nil, err
	case result := <-resCh:
		return &pb.RespJob{Status: result.Status, Message: result.Message}, nil
	case <-time.After(APITimeOut):
		return nil, errors.New("Time out")
	}
}
