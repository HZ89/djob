// Code generated by protoc-gen-go.
// source: job.proto
// DO NOT EDIT!

/*
Package message is a generated protocol buffer package.

It is generated from these files:
	job.proto
	serfQueryParams.proto

It has these top-level messages:
	Name
	Job
	JobStatus
	Execution
	Result
	Resp
	NewJobQueryParams
	RunJobQueryParams
*/
package message

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Name struct {
	Name string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
}

func (m *Name) Reset()                    { *m = Name{} }
func (m *Name) String() string            { return proto.CompactTextString(m) }
func (*Name) ProtoMessage()               {}
func (*Name) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Name) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type Job struct {
	Name               string   `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	Region             string   `protobuf:"bytes,2,opt,name=Region" json:"Region,omitempty"`
	Schedule           string   `protobuf:"bytes,3,opt,name=schedule" json:"schedule,omitempty"`
	Shell              bool     `protobuf:"varint,4,opt,name=Shell" json:"Shell,omitempty"`
	Command            string   `protobuf:"bytes,5,opt,name=Command" json:"Command,omitempty"`
	Expression         string   `protobuf:"bytes,6,opt,name=Expression" json:"Expression,omitempty"`
	Retries            int64    `protobuf:"varint,7,opt,name=Retries" json:"Retries,omitempty"`
	BeingDependentJobs []string `protobuf:"bytes,8,rep,name=BeingDependentJobs" json:"BeingDependentJobs,omitempty"`
	ParentJob          string   `protobuf:"bytes,9,opt,name=ParentJob" json:"ParentJob,omitempty"`
	Parallel           bool     `protobuf:"varint,10,opt,name=Parallel" json:"Parallel,omitempty"`
	Concurrent         bool     `protobuf:"varint,11,opt,name=Concurrent" json:"Concurrent,omitempty"`
	Disposable         bool     `protobuf:"varint,12,opt,name=Disposable" json:"Disposable,omitempty"`
	SchedulerNodeName  string   `protobuf:"bytes,13,opt,name=SchedulerNodeName" json:"SchedulerNodeName,omitempty"`
}

func (m *Job) Reset()                    { *m = Job{} }
func (m *Job) String() string            { return proto.CompactTextString(m) }
func (*Job) ProtoMessage()               {}
func (*Job) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Job) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Job) GetRegion() string {
	if m != nil {
		return m.Region
	}
	return ""
}

func (m *Job) GetSchedule() string {
	if m != nil {
		return m.Schedule
	}
	return ""
}

func (m *Job) GetShell() bool {
	if m != nil {
		return m.Shell
	}
	return false
}

func (m *Job) GetCommand() string {
	if m != nil {
		return m.Command
	}
	return ""
}

func (m *Job) GetExpression() string {
	if m != nil {
		return m.Expression
	}
	return ""
}

func (m *Job) GetRetries() int64 {
	if m != nil {
		return m.Retries
	}
	return 0
}

func (m *Job) GetBeingDependentJobs() []string {
	if m != nil {
		return m.BeingDependentJobs
	}
	return nil
}

func (m *Job) GetParentJob() string {
	if m != nil {
		return m.ParentJob
	}
	return ""
}

func (m *Job) GetParallel() bool {
	if m != nil {
		return m.Parallel
	}
	return false
}

func (m *Job) GetConcurrent() bool {
	if m != nil {
		return m.Concurrent
	}
	return false
}

func (m *Job) GetDisposable() bool {
	if m != nil {
		return m.Disposable
	}
	return false
}

func (m *Job) GetSchedulerNodeName() string {
	if m != nil {
		return m.SchedulerNodeName
	}
	return ""
}

type JobStatus struct {
	Name            string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	SuccessCount    int64  `protobuf:"varint,2,opt,name=SuccessCount" json:"SuccessCount,omitempty"`
	ErrorCount      int64  `protobuf:"varint,3,opt,name=ErrorCount" json:"ErrorCount,omitempty"`
	LastHandleAgent string `protobuf:"bytes,4,opt,name=LastHandleAgent" json:"LastHandleAgent,omitempty"`
	LastSuccess     string `protobuf:"bytes,5,opt,name=LastSuccess" json:"LastSuccess,omitempty"`
	LastError       string `protobuf:"bytes,6,opt,name=LastError" json:"LastError,omitempty"`
}

func (m *JobStatus) Reset()                    { *m = JobStatus{} }
func (m *JobStatus) String() string            { return proto.CompactTextString(m) }
func (*JobStatus) ProtoMessage()               {}
func (*JobStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *JobStatus) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *JobStatus) GetSuccessCount() int64 {
	if m != nil {
		return m.SuccessCount
	}
	return 0
}

func (m *JobStatus) GetErrorCount() int64 {
	if m != nil {
		return m.ErrorCount
	}
	return 0
}

func (m *JobStatus) GetLastHandleAgent() string {
	if m != nil {
		return m.LastHandleAgent
	}
	return ""
}

func (m *JobStatus) GetLastSuccess() string {
	if m != nil {
		return m.LastSuccess
	}
	return ""
}

func (m *JobStatus) GetLastError() string {
	if m != nil {
		return m.LastError
	}
	return ""
}

type Execution struct {
	Name       string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	Cmd        string `protobuf:"bytes,2,opt,name=Cmd" json:"Cmd,omitempty"`
	Output     []byte `protobuf:"bytes,3,opt,name=Output,proto3" json:"Output,omitempty"`
	Succeed    bool   `protobuf:"varint,4,opt,name=Succeed" json:"Succeed,omitempty"`
	StartTime  string `protobuf:"bytes,5,opt,name=StartTime" json:"StartTime,omitempty"`
	FinishTime string `protobuf:"bytes,6,opt,name=FinishTime" json:"FinishTime,omitempty"`
	NodeName   string `protobuf:"bytes,7,opt,name=NodeName" json:"NodeName,omitempty"`
	JobName    string `protobuf:"bytes,8,opt,name=JobName" json:"JobName,omitempty"`
}

func (m *Execution) Reset()                    { *m = Execution{} }
func (m *Execution) String() string            { return proto.CompactTextString(m) }
func (*Execution) ProtoMessage()               {}
func (*Execution) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Execution) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Execution) GetCmd() string {
	if m != nil {
		return m.Cmd
	}
	return ""
}

func (m *Execution) GetOutput() []byte {
	if m != nil {
		return m.Output
	}
	return nil
}

func (m *Execution) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *Execution) GetStartTime() string {
	if m != nil {
		return m.StartTime
	}
	return ""
}

func (m *Execution) GetFinishTime() string {
	if m != nil {
		return m.FinishTime
	}
	return ""
}

func (m *Execution) GetNodeName() string {
	if m != nil {
		return m.NodeName
	}
	return ""
}

func (m *Execution) GetJobName() string {
	if m != nil {
		return m.JobName
	}
	return ""
}

type Result struct {
	Err bool `protobuf:"varint,1,opt,name=err" json:"err,omitempty"`
}

func (m *Result) Reset()                    { *m = Result{} }
func (m *Result) String() string            { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()               {}
func (*Result) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Result) GetErr() bool {
	if m != nil {
		return m.Err
	}
	return false
}

type Resp struct {
	Error string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
	Data  []*Job `protobuf:"bytes,2,rep,name=data" json:"data,omitempty"`
}

func (m *Resp) Reset()                    { *m = Resp{} }
func (m *Resp) String() string            { return proto.CompactTextString(m) }
func (*Resp) ProtoMessage()               {}
func (*Resp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Resp) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *Resp) GetData() []*Job {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*Name)(nil), "message.Name")
	proto.RegisterType((*Job)(nil), "message.Job")
	proto.RegisterType((*JobStatus)(nil), "message.JobStatus")
	proto.RegisterType((*Execution)(nil), "message.Execution")
	proto.RegisterType((*Result)(nil), "message.Result")
	proto.RegisterType((*Resp)(nil), "message.Resp")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Job service

type JobClient interface {
	GetJob(ctx context.Context, in *Name, opts ...grpc.CallOption) (*Job, error)
	GetExecution(ctx context.Context, in *Name, opts ...grpc.CallOption) (*Execution, error)
	ExecDone(ctx context.Context, in *Execution, opts ...grpc.CallOption) (*Result, error)
}

type jobClient struct {
	cc *grpc.ClientConn
}

func NewJobClient(cc *grpc.ClientConn) JobClient {
	return &jobClient{cc}
}

func (c *jobClient) GetJob(ctx context.Context, in *Name, opts ...grpc.CallOption) (*Job, error) {
	out := new(Job)
	err := grpc.Invoke(ctx, "/message.job/GetJob", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobClient) GetExecution(ctx context.Context, in *Name, opts ...grpc.CallOption) (*Execution, error) {
	out := new(Execution)
	err := grpc.Invoke(ctx, "/message.job/GetExecution", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobClient) ExecDone(ctx context.Context, in *Execution, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/message.job/ExecDone", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Job service

type JobServer interface {
	GetJob(context.Context, *Name) (*Job, error)
	GetExecution(context.Context, *Name) (*Execution, error)
	ExecDone(context.Context, *Execution) (*Result, error)
}

func RegisterJobServer(s *grpc.Server, srv JobServer) {
	s.RegisterService(&_Job_serviceDesc, srv)
}

func _Job_GetJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Name)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServer).GetJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message.job/GetJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServer).GetJob(ctx, req.(*Name))
	}
	return interceptor(ctx, in, info, handler)
}

func _Job_GetExecution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Name)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServer).GetExecution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message.job/GetExecution",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServer).GetExecution(ctx, req.(*Name))
	}
	return interceptor(ctx, in, info, handler)
}

func _Job_ExecDone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Execution)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServer).ExecDone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message.job/ExecDone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServer).ExecDone(ctx, req.(*Execution))
	}
	return interceptor(ctx, in, info, handler)
}

var _Job_serviceDesc = grpc.ServiceDesc{
	ServiceName: "message.job",
	HandlerType: (*JobServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetJob",
			Handler:    _Job_GetJob_Handler,
		},
		{
			MethodName: "GetExecution",
			Handler:    _Job_GetExecution_Handler,
		},
		{
			MethodName: "ExecDone",
			Handler:    _Job_ExecDone_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "job.proto",
}

func init() { proto.RegisterFile("job.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 554 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x54, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x6d, 0xea, 0x34, 0x71, 0xa6, 0xa9, 0x0a, 0x2b, 0x84, 0x56, 0x11, 0x42, 0x95, 0x2f, 0xe4,
	0x80, 0x22, 0xd1, 0xde, 0x91, 0x20, 0x29, 0x20, 0x84, 0x0a, 0x72, 0xf8, 0x01, 0x27, 0x5e, 0x25,
	0x46, 0x8e, 0xd7, 0xda, 0x5d, 0x4b, 0xfd, 0x0b, 0x6e, 0xfc, 0x15, 0x67, 0x7e, 0x87, 0x99, 0xd9,
	0x8d, 0x9d, 0xb6, 0x39, 0xc5, 0xef, 0xbd, 0xd9, 0x9d, 0x99, 0x37, 0xb3, 0x81, 0xd1, 0x2f, 0xbd,
	0x9a, 0xd5, 0x46, 0x3b, 0x2d, 0x86, 0x3b, 0x65, 0x6d, 0xb6, 0x51, 0xc9, 0x04, 0xfa, 0x77, 0xd9,
	0x4e, 0x09, 0xe1, 0x7f, 0x65, 0xef, 0xaa, 0x37, 0x1d, 0xa5, 0xfc, 0x9d, 0xfc, 0x8e, 0x20, 0xfa,
	0xaa, 0x57, 0xc7, 0x34, 0xf1, 0x12, 0x06, 0xa9, 0xda, 0x14, 0xba, 0x92, 0xa7, 0xcc, 0x06, 0x24,
	0x26, 0x10, 0xdb, 0xf5, 0x56, 0xe5, 0x4d, 0xa9, 0x64, 0xc4, 0x4a, 0x8b, 0xc5, 0x0b, 0x38, 0x5b,
	0x6e, 0x55, 0x59, 0xca, 0x3e, 0x0a, 0x71, 0xea, 0x81, 0x90, 0x30, 0x9c, 0xeb, 0xdd, 0x2e, 0xab,
	0x72, 0x79, 0xc6, 0x07, 0xf6, 0x50, 0xbc, 0x06, 0xb8, 0xbd, 0xaf, 0x0d, 0x56, 0x4a, 0x79, 0x06,
	0x2c, 0x1e, 0x30, 0x74, 0x32, 0x55, 0xce, 0x14, 0xca, 0xca, 0x21, 0x8a, 0x51, 0xba, 0x87, 0x62,
	0x06, 0xe2, 0xa3, 0x2a, 0xaa, 0xcd, 0x42, 0xd5, 0xaa, 0xca, 0x55, 0xe5, 0xb0, 0x0d, 0x2b, 0xe3,
	0xab, 0x08, 0x6f, 0x38, 0xa2, 0x88, 0x57, 0x30, 0xfa, 0x91, 0x19, 0x8f, 0xe4, 0x88, 0x13, 0x75,
	0x04, 0xf5, 0x84, 0x20, 0x2b, 0x4b, 0x55, 0x4a, 0xe0, 0xd2, 0x5b, 0x4c, 0x35, 0xce, 0x75, 0xb5,
	0x6e, 0x0c, 0x05, 0xcb, 0x73, 0x56, 0x0f, 0x18, 0xd2, 0x17, 0x85, 0xad, 0xb5, 0xcd, 0x56, 0xe8,
	0xc8, 0xd8, 0xeb, 0x1d, 0x23, 0xde, 0xc2, 0xf3, 0x65, 0xf0, 0xc7, 0xdc, 0xe9, 0x5c, 0xb1, 0xd1,
	0x17, 0x5c, 0xc1, 0x53, 0x21, 0xf9, 0xdb, 0x83, 0x11, 0x56, 0xb4, 0x74, 0x99, 0x6b, 0xec, 0xd1,
	0xb9, 0x24, 0x30, 0x5e, 0x36, 0xeb, 0x35, 0x3a, 0x34, 0xd7, 0x0d, 0x56, 0x74, 0xca, 0xc6, 0x3c,
	0xe0, 0xd8, 0x57, 0x63, 0xb4, 0xf1, 0x11, 0x11, 0x47, 0x1c, 0x30, 0x62, 0x0a, 0x97, 0xdf, 0x32,
	0xeb, 0xbe, 0xe0, 0x0c, 0x4a, 0xf5, 0x61, 0x43, 0x8d, 0xf5, 0x39, 0xc5, 0x63, 0x5a, 0x5c, 0xc1,
	0x39, 0x51, 0xe1, 0xf6, 0x30, 0xbf, 0x43, 0x8a, 0x9c, 0x25, 0xc8, 0xb7, 0x87, 0x11, 0x76, 0x44,
	0xf2, 0x0f, 0xfb, 0xb9, 0xbd, 0x57, 0xeb, 0xc6, 0xd1, 0x3c, 0x8f, 0xf5, 0xf3, 0x0c, 0xa2, 0xf9,
	0x2e, 0x0f, 0x4b, 0x46, 0x9f, 0xb4, 0x79, 0xdf, 0x1b, 0x57, 0x37, 0xbe, 0xf2, 0x71, 0x1a, 0x10,
	0x6d, 0x03, 0x27, 0x55, 0x79, 0xd8, 0xaf, 0x3d, 0xa4, 0x1a, 0xd0, 0x31, 0xe3, 0x7e, 0x16, 0x78,
	0xb9, 0xaf, 0xb1, 0x23, 0xc8, 0x8d, 0x4f, 0x45, 0x55, 0xd8, 0x2d, 0xcb, 0x61, 0xcb, 0x3a, 0x86,
	0xa6, 0xdf, 0x0e, 0x66, 0xe8, 0x37, 0x7a, 0x8f, 0x29, 0x27, 0x8e, 0x83, 0xa5, 0xd8, 0xef, 0x6e,
	0x80, 0xf8, 0xae, 0xf0, 0x45, 0xd8, 0xa6, 0x74, 0xd4, 0x81, 0x32, 0x86, 0x9b, 0x8a, 0x53, 0xfa,
	0x4c, 0xde, 0x43, 0x1f, 0xb5, 0x9a, 0xde, 0x83, 0x62, 0x5f, 0x7c, 0xc3, 0x1e, 0xa0, 0xa7, 0xfd,
	0x3c, 0x73, 0x19, 0xb6, 0x1c, 0x4d, 0xcf, 0xaf, 0xc7, 0xb3, 0xf0, 0x52, 0x67, 0x78, 0x73, 0xca,
	0xca, 0xf5, 0x9f, 0x1e, 0x44, 0xf8, 0x94, 0xc5, 0x1b, 0x18, 0x7c, 0x56, 0xbc, 0xa1, 0x17, 0x6d,
	0x14, 0x25, 0x9f, 0x3c, 0x38, 0x94, 0x9c, 0x88, 0x1b, 0x18, 0x63, 0x60, 0x67, 0xf4, 0xa3, 0x70,
	0xd1, 0xc2, 0x36, 0x04, 0x0f, 0xbd, 0x83, 0x98, 0xe0, 0x42, 0x57, 0xf8, 0xef, 0xf0, 0x34, 0x62,
	0x72, 0xd9, 0x72, 0xbe, 0xd1, 0xe4, 0x64, 0x35, 0xe0, 0x3f, 0x97, 0x9b, 0xff, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x22, 0x03, 0x8c, 0xb3, 0x69, 0x04, 0x00, 0x00,
}
