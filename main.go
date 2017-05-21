package main

import (
	pj "local/djob/job"
	"local/djob/djob"
	"local/djob/rpc"
	"sync"
)

type MockDServer struct {}

func (s *MockDServer)JobInfo(jobName string) (*pj.Job, error) {
	return &pj.Job{
		Name: jobName,
		SchedulerName: "RPC_TEST",
	}, nil
}

func (s *MockDServer)ExecDone(execution *pj.Execution) error {
	djob.Log.Infof("get execution: %s", execution.Name)
	return nil
}

func main() {
	var djs MockDServer
	rpcserver := rpc.NewRPCserver("127.0.0.1", 8866, false, &djs, "", "")


	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		rpcserver.Run()
		wg.Done()
	}()

	rpcclient, err := rpc.NewRpcClient("127.0.0.1", 8866, false, "", "")
	if err != nil {
		djob.Log.Fatal(err)
	}
	job, err := rpcclient.GotJob("testA")
	if err != nil {
		djob.Log.Error(err)
	}
	djob.Log.Infof("get job: jobName %s", job.Name)
	rpcclient.Shutdown()

}