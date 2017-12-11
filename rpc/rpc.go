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
	"strings"
	"time"

	"github.com/HZ89/djob/log"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/HZ89/djob/errors"
	pb "github.com/HZ89/djob/message"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

var registry = make(map[string]reflect.Type)

func init() {
	registry["message.Job"] = reflect.TypeOf(pb.Job{})
	registry["message.Execution"] = reflect.TypeOf(pb.Execution{})
	registry["message.JobStatus"] = reflect.TypeOf(pb.JobStatus{})
}

type Operator interface {
	// use to client get job
	GetJob(name, region string) (*pb.Job, error)
	// client send execution to server
	SendBackExecution(execution *pb.Execution) error
	// forwarding CRUD operation between servers
	PerformOps(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int32, error)
	// forwarding run job action between servers
	RunJob(name, region string) (*pb.Execution, error)
	//
}

type RpcServer struct {
	bindIp    string       // rpc server bind ip
	rpcPort   int          // rpc server bind port
	tlsopt    *TlsOpt      // tls options
	operator  Operator     // operator interface
	rpcServer *grpc.Server // grpc server
}

type TlsOpt struct {
	CertFile   string // cert file path
	KeyFile    string // private key file path
	CaFile     string // client use this root ca
	ServerHost string // the server name use to verify the hostname returned by TLS handshake
}

func NewRPCServer(bindIp string, port int, operator Operator, tlsopt *TlsOpt) *RpcServer {
	return &RpcServer{
		bindIp:   bindIp,
		rpcPort:  port,
		operator: operator,
		tlsopt:   tlsopt,
	}
}

// AcquireToken issue token to agent, if have no token left, give agent a wait time duration
func (s *RpcServer) AcquireToken(ctx context.Context, req *pb.TokenReqMessage) (*pb.TokenRespMessage, error) {

}

// RecedeToken agent recede a token
func (s *RpcServer) RecedeToken(ctx context.Context, req *pb.TokenReqMessage) (*pb.TokenRespMessage, error) {

}

// forwarding run job action
func (s *RpcServer) ProxyJobRun(ctx context.Context, in *pb.Job) (*pb.Execution, error) {
	exec, err := s.operator.RunJob(in.Name, in.Region)
	if err != nil {
		return nil, errors.GenGRPCErr(err)
	}
	return exec, nil
}

// get job info send to client
func (s *RpcServer) GetJob(ctx context.Context, job *pb.Job) (*pb.Job, error) {
	job, err := s.operator.GetJob(job.Name, job.Region)
	if err != nil {
		return nil, errors.GenGRPCErr(err)
	}
	return job, nil
}

// receive the execution result
func (s *RpcServer) ExecDone(ctx context.Context, execution *pb.Execution) (*google_protobuf.Empty, error) {
	if err := s.operator.SendBackExecution(execution); err != nil {
		return nil, errors.GenGRPCErr(err)
	}
	return &google_protobuf.Empty{}, nil
}

// receive job CRUD ops
func (s *RpcServer) DoOps(ctx context.Context, p *pb.Params) (*pb.Result, error) {
	class := strings.Split(p.Obj.TypeUrl, "/")[1]
	log.FmdLoger.WithField("type", class).Debug("RPC: Server got a class")
	t, ok := registry[class]
	if !ok {
		return nil, errors.ErrType.GenGRPCErr()
	}
	instance := reflect.New(t).Interface()
	log.FmdLoger.WithField("instance_type", reflect.TypeOf(instance)).Debug("RPC: Server prepare use instance decode obj")
	if err := ptypes.UnmarshalAny(p.Obj, instance.(proto.Message)); err != nil {
		return nil, errors.GenGRPCErr(err)
	}
	r, count, err := s.operator.PerformOps(instance, p.Ops, p.Search)
	if err != nil {
		return nil, errors.GenGRPCErr(err)
	}
	var rs []*any.Any
	for _, i := range r {
		b, err := ptypes.MarshalAny(i.(proto.Message))
		if err != nil {
			return nil, errors.GenGRPCErr(err)
		}
		rs = append(rs, b)
	}
	return &pb.Result{
		Succeed:    true,
		Objs:       rs,
		MaxPageNum: count,
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

// start grpc server process
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

// shutdown grpc server
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
	serverIp   string           // rpc server ip
	serverPort int              // rpc server port
	tlsopt     *TlsOpt          // tls options
	client     pb.JobClient     // grpc client interface
	conn       *grpc.ClientConn // grpc client connection to an RPC server
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
	return c.conn.Close()
}

func (c *RpcClient) ProxyJobRun(name, region string) (*pb.Execution, error) {
	exec, err := c.client.ProxyJobRun(context.Background(), &pb.Job{Name: name, Region: region})
	if err != nil {
		if terr, ok := errors.NewFromGRPCErr(err); ok {
			return nil, terr
		}
		log.FmdLoger.WithError(err).Error("RPC-Client: got err but not a GRPC error")
		return nil, err
	}
	return exec, nil
}

func (c *RpcClient) GetJob(name, region string) (*pb.Job, error) {
	p := pb.Job{Name: name, Region: region}
	job, err := c.client.GetJob(context.Background(), &p)
	if err != nil {
		if terr, ok := errors.NewFromGRPCErr(err); ok {
			return nil, terr
		}
		log.FmdLoger.WithError(err).Error("RPC-Client: got err but not a GRPC error")
		return nil, err
	}

	return job, nil
}

func (c *RpcClient) ExecDone(execution *pb.Execution) error {
	_, err := c.client.ExecDone(context.Background(), execution)
	if err != nil {
		if terr, ok := errors.NewFromGRPCErr(err); ok {
			return terr
		}
		log.FmdLoger.WithError(err).Error("RPC-Client: got err but not a GRPC error")
		return err
	}
	return nil
}

func (c *RpcClient) DoOps(obj interface{}, ops pb.Ops, search *pb.Search) (instances []interface{}, count int32, err error) {
	pbObj, err := ptypes.MarshalAny(obj.(proto.Message))
	if err != nil {
		return nil, 0, err
	}
	log.FmdLoger.WithField("Any Obj", pbObj).Debugf("RPC: client prepare DoOps rpc call, Obj: %v", pbObj)
	r, err := c.client.DoOps(context.Background(), &pb.Params{Obj: pbObj, Ops: ops, Search: search})
	if err != nil {
		if terr, ok := errors.NewFromGRPCErr(err); ok {
			return nil, 0, terr
		}
		log.FmdLoger.WithError(err).Error("RPC-Client: got err but not a GRPC error")
		return nil, 0, err
	}
	log.FmdLoger.WithField("obj", r).Debug("RPC: RPC client call DoOps done, got this")
	count = r.MaxPageNum
	for _, o := range r.Objs {
		t, ok := registry[strings.Split(o.TypeUrl, "/")[1]]
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
