package rpc

import (
	pb "local/djob/message"
	"local/djob/job"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"net"
	"fmt"
	"google.golang.org/grpc/credentials"
	"golang.org/x/net/context"
	"time"
)

type DjobServer interface {
	JobInfo(jobName string) (*job.Job, error)
	ExecDone(execution *job.Execution) error
}

type RpcServer struct {
	bindIp    string
	rpcPort   int
	tls       bool
	certFile  string
	keyFile   string
	dserver   DjobServer
	rpcServer *grpc.Server
}

func NewRPCserver(bindIp string, port int, tls bool, server DjobServer, certFile string, keyFile string) *RpcServer {
	return &RpcServer{
		bindIp:   bindIp,
		rpcPort:  port,
		tls:      tls,
		dserver:  server,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func (s *RpcServer) GetJob(ctx context.Context, name *pb.Name) (*pb.Job, error) {
	jobInfo, err := s.dserver.JobInfo(name.JobName)
	if err != nil {
		return nil, err
	}

	pjob := &pb.Job{
		JobName: jobInfo.Name,
	}

	return pjob, nil
}

func (s *RpcServer) ExecDone(ctx context.Context, execution *pb.Execution) (*pb.Result, error) {
	stime, _ := ptypes.Timestamp(execution.StartTime)
	ftime, _ := ptypes.Timestamp(execution.FinishTime)
	et := job.Execution{
		Name:       execution.Name,
		Cmd:        execution.Cmd,
		Output:     []byte(execution.Output),
		StartTime:  stime,
		FinishTime: ftime,
	}
	if err := s.dserver.ExecDone(&et); err != nil {
		return &pb.Result{
			Err: false,
		}, nil
	}
	return &pb.Result{
		Err: true,
	}, nil
}

func (s *RpcServer) listen() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.bindIp, s.rpcPort))
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption

	if s.tls {
		creds, err := credentials.NewServerTLSFromFile(s.certFile, s.keyFile)
		if err != nil {
			return err
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	s.rpcServer = grpc.NewServer(opts...)
	pb.RegisterJobServer(s.rpcServer, s)

	// grpc server run in a single goroutine
	go s.rpcServer.Serve(lis)

	return nil
}

func (s *RpcServer) Run() error {
	if err := s.listen(); err != nil {
		return err
	}
	return nil
}

func (s *RpcServer) Shutdown(timeout time.Duration) error {
	timeoutCh := time.After(timeout)
	var shutdownCh chan error

	go func() {
		s.rpcServer.GracefulStop()
		shutdownCh <- nil
	}()

	select {

	case err := <-shutdownCh:
		return err
	case <-timeoutCh:
		s.rpcServer.Stop()
		return nil

	}

}

type RpcClient struct {
	serverIp   string
	serverPort int
	withTls    bool
	caFile     string
	serverHost string //The server name use to verify the hostname returned by TLS handshake
	client     pb.JobClient
	conn       *grpc.ClientConn
}

func NewRpcClient(serveraddr string, serverport int, tls bool, cafile string, serverhost string) (*RpcClient, error) {
	var opts []grpc.DialOption
	client := &RpcClient{
		serverIp:   serveraddr,
		serverPort: serverport,
		withTls:    tls,
		caFile:     cafile,
		serverHost: serverhost,
	}
	if client.withTls {
		var sn string
		if client.serverHost != "" {
			sn = client.serverHost
		}
		var creds credentials.TransportCredentials
		if client.caFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(client.caFile, sn)
			if err != nil {
				return nil, err
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, sn)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	var err error
	client.conn, err = grpc.Dial(fmt.Sprintf("%s:%d", client.serverIp, client.serverPort), opts...)
	if err != nil {
		return nil, err
	}
	client.client = pb.NewJobClient(client.conn)
	return client, nil
}

func (c *RpcClient) Shutdown() error {
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *RpcClient)GotJob(jobName string) (*job.Job, error)  {
	pbName := pb.Name{JobName:jobName}
	pbjob, err := c.client.GetJob(context.Background(), &pbName)
	if err != nil {
		return nil, err
	}

	return &job.Job{
		Name: pbjob.JobName,
	}, nil
}

func (c *RpcClient)ExecDone(execution *job.Execution) (bool, error) {
	pstime, _ := ptypes.TimestampProto(execution.StartTime)
	pftime, _ := ptypes.TimestampProto(execution.FinishTime)
	pbexecution := &pb.Execution{
		Name: execution.Name,
		Cmd: execution.Cmd,
		Output: string(execution.Output),
		StartTime: pstime,
		FinishTime: pftime,
		Succeed: execution.Succeed,
	}

	pbresutl, err := c.client.ExecDone(context.Background(), pbexecution)
	if err != nil {
		return false, err
	}
	return pbresutl.Err, nil
}
