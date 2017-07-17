package rpc

import (
	"fmt"
	//	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "local/djob/message"
	"net"
	"time"
)

type DjobServer interface {
	JobInfo(jobName string) (*pb.Job, error)
	ExecDone(execution *pb.Execution) error
}

type RpcServer struct {
	bindIp    string
	rpcPort   int
	tlsopt    *TlsOpt
	dserver   DjobServer
	rpcServer *grpc.Server
}

type TlsOpt struct {
	CertFile   string
	KeyFile    string
	CaFile     string
	ServerHost string //The server name use to verify the hostname returned by TLS handshake
}

func NewRPCserver(bindIp string, port int, server DjobServer, tlsopt *TlsOpt) *RpcServer {
	return &RpcServer{
		bindIp:  bindIp,
		rpcPort: port,
		dserver: server,
		tlsopt:  tlsopt,
	}
}

func (s *RpcServer) GetJob(ctx context.Context, name *pb.Name) (*pb.Job, error) {
	jobInfo, err := s.dserver.JobInfo(name.JobName)
	if err != nil {
		return nil, err
	}

	pjob := &pb.Job{
		Name: jobInfo.Name,
	}

	return pjob, nil
}

func (s *RpcServer) ExecDone(ctx context.Context, execution *pb.Execution) (*pb.Result, error) {
	if err := s.dserver.ExecDone(execution); err != nil {
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

	if s.tlsopt {
		creds, err := credentials.NewServerTLSFromFile(s.tlsopt.CaFile, s.tlsopt.KeyFile)
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
	tlsopt     *TlsOpt
	client     pb.JobClient
	conn       *grpc.ClientConn
}

func NewRpcClient(serveraddr string, serverport int, tlsopt *TlsOpt) (*RpcClient, error) {
	var opts []grpc.DialOption
	client := &RpcClient{
		serverIp:   serveraddr,
		serverPort: serverport,
		tlsopt:     tlsopt,
	}
	if client.tlsopt {
		var sn string
		if client.tlsopt.ServerHost != "" {
			sn = client.tlsopt.ServerHost
		}
		var creds credentials.TransportCredentials
		if client.tlsopt.CaFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(client.tlsopt.CaFile, sn)
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

func (c *RpcClient) GotJob(jobName string) (*pb.Job, error) {
	pbName := pb.Name{JobName: jobName}
	pbjob, err := c.client.GetJob(context.Background(), &pbName)
	if err != nil {
		return nil, err
	}

	return pbjob, nil
}

func (c *RpcClient) ExecDone(execution *pb.Execution) (bool, error) {

	pbresutl, err := c.client.ExecDone(context.Background(), execution)
	if err != nil {
		return false, err
	}
	return pbresutl.Err, nil
}
