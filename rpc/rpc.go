/*
 * Copyright (c) 2017.  Harrison Zhu <wcg6121@gmail.com>
 * This file is part of djob <https://github.com/HZ89/djob>.
 *
 * djob is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * djob is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with djob.  If not, see <http://www.gnu.org/licenses/>.
 */

package rpc

import (
	"fmt"
	"net"
	"reflect"
	"time"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"version.uuzu.com/zhuhuipeng/djob/errors"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

var registry = make(map[string]reflect.Type)

func init() {
	registry["Job"] = reflect.TypeOf(&pb.Job{})
	registry["Execution"] = reflect.TypeOf(&pb.Execution{})
	registry["JobStatus"] = reflect.TypeOf(&pb.JobStatus{})
}

type Operator interface {
	GetJob(name, region string) (*pb.Job, error)
	SendBackExecution(execution *pb.Execution) error
	PerformOps(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int, error)
	RunJob(name, region string) (*pb.Execution, error)
}

type RpcServer struct {
	bindIp    string
	rpcPort   int
	tlsopt    *TlsOpt
	operator  Operator
	rpcServer *grpc.Server
}

type TlsOpt struct {
	CertFile   string
	KeyFile    string
	CaFile     string // Client use this root ca
	ServerHost string // The server name use to verify the hostname returned by TLS handshake
}

func NewRPCServer(bindIp string, port int, operator Operator, tlsopt *TlsOpt) *RpcServer {
	return &RpcServer{
		bindIp:   bindIp,
		rpcPort:  port,
		operator: operator,
		tlsopt:   tlsopt,
	}
}

func (s *RpcServer) ProxyJobRun(ctx context.Context, in *pb.Job) (*pb.Execution, error) {
	exec, err := s.operator.RunJob(in.Name, in.Region)
	if err != nil {
		return nil, err
	}
	return exec, nil
}

func (s *RpcServer) GetJob(ctx context.Context, job *pb.Job) (*pb.Job, error) {
	job, err := s.operator.GetJob(job.Name, job.Region)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *RpcServer) ExecDone(ctx context.Context, execution *pb.Execution) (*google_protobuf.Empty, error) {
	if err := s.operator.SendBackExecution(execution); err != nil {
		return nil, err
	}
	return &google_protobuf.Empty{}, nil
}

func (s *RpcServer) DoOps(ctx context.Context, p *pb.Params) (*pb.Result, error) {
	class := p.Obj.TypeUrl
	t, ok := registry[class]
	if !ok {
		return nil, errors.ErrType
	}
	instance := reflect.New(t).Interface()
	if err := ptypes.UnmarshalAny(p.Obj, instance.(proto.Message)); err != nil {
		return nil, err
	}
	r, count, err := s.operator.PerformOps(instance, p.Ops, p.Search)
	if err != nil {
		return nil, err
	}
	var rs []*any.Any
	for _, i := range r {
		b, err := ptypes.MarshalAny(i.(proto.Message))
		if err != nil {
			return nil, err
		}
		rs = append(rs, b)
	}
	return &pb.Result{
		Succeed:    true,
		Objs:       rs,
		MaxPageNum: int32(count),
	}, nil

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
	errCh := make(chan error, 1)
	doneCh := make(chan struct{}, 1)
	go func() {
		if err := s.listen(); err != nil {
			errCh <- err
		}
		doneCh <- struct{}{}
	}()
	select {
	case err := <-errCh:
		return err
	case <-doneCh:
		return nil
	}
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

func (c *RpcClient) ProxyJobRun(name, region string) (*pb.Execution, error) {
	exec, err := c.client.ProxyJobRun(context.Background(), &pb.Job{Name: name, Region: region})
	if err != nil {
		return nil, err
	}
	return exec, nil
}

func (c *RpcClient) GetJob(name, region string) (*pb.Job, error) {
	p := pb.Job{Name: name, Region: region}
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

func (c *RpcClient) DoOps(obj interface{}, ops pb.Ops, search *pb.Search) (instances []interface{}, count int, err error) {
	pbObj, err := ptypes.MarshalAny(obj.(proto.Message))
	if err != nil {
		return nil, 0, err
	}
	r, err := c.client.DoOps(context.Background(), &pb.Params{Obj: pbObj, Ops: ops, Search: search})
	if err != nil {
		return nil, 0, err
	}
	count = int(r.MaxPageNum)
	for _, o := range r.Objs {
		t, ok := registry[o.TypeUrl]
		if !ok {
			return nil, 0, errors.ErrType
		}
		instance := reflect.New(t).Interface()
		if err = ptypes.UnmarshalAny(o, instance.(proto.Message)); err != nil {
			return
		}
		instances = append(instances, instance)
	}
	return
}
