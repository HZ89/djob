package rpc

import (
	"fmt"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"time"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

type DjobServer interface {
	JobInfo(name, region string) (*pb.Job, error)
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
	CaFile     string // Client use this root ca
	ServerHost string // The server name use to verify the hostname returned by TLS handshake
}

func NewRPCServer(bindIp string, port int, server DjobServer, tlsopt *TlsOpt) *RpcServer {
	return &RpcServer{
		bindIp:  bindIp,
		rpcPort: port,
		dserver: server,
		tlsopt:  tlsopt,
	}
}

func (s *RpcServer) GetJob(ctx context.Context, params *pb.Params) (*pb.Job, error) {
	jobInfo, err := s.dserver.JobInfo(params.Name, params.Region)
	if err != nil {
		return nil, err
	}

	job := &pb.Job{
		Name: jobInfo.Name,
	}

	return job, nil
}

func (s *RpcServer) ExecDone(ctx context.Context, execution *pb.Execution) (*google_protobuf.Empty, error) {
	if err := s.dserver.ExecDone(execution); err != nil {
		return nil, err
	}
	return &google_protobuf.Empty{}, nil
}

func (s *RpcServer) listen() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.bindIp, s.rpcPort))
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption

	if s.tlsopt != nil {
		creds, err := credentials.NewServerTLSFromFile(s.tlsopt.CertFile, s.tlsopt.KeyFile)
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
	if client.tlsopt != nil {
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

func (c *RpcClient) GetJob(name, region string) (*pb.Job, error) {
	p := pb.Params{Name: name, Region: region}
	job, err := c.client.GetJob(context.Background(), &p)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (c *RpcClient) ExecDone(execution *pb.Execution) error {
	_, err := c.client.ExecDone(context.Background(), execution)
	if err != nil {
		return err
	}
	return nil
}
