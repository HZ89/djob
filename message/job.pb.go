// Code generated by protoc-gen-go.
// source: message/job.proto
// DO NOT EDIT!

package message

import "github.com/golang/protobuf/proto"
import "fmt"
import "math"
import google_protobuf "github.com/golang/protobuf/ptypes/empty"
import google_protobuf1 "github.com/golang/protobuf/ptypes/any"

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Ops int32

const (
	Ops_READ   Ops = 0
	Ops_ADD    Ops = 1
	Ops_MODIFY Ops = 2
	Ops_DELETE Ops = 3
)

var Ops_name = map[int32]string{
	0: "READ",
	1: "ADD",
	2: "MODIFY",
	3: "DELETE",
}
var Ops_value = map[string]int32{
	"READ":   0,
	"ADD":    1,
	"MODIFY": 2,
	"DELETE": 3,
}

func (x Ops) String() string {
	return proto.EnumName(Ops_name, int32(x))
}
func (Ops) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

type Search struct {
	Conditions []string `protobuf:"bytes,1,rep,name=Conditions" json:"Conditions,omitempty"`
	Links      []string `protobuf:"bytes,2,rep,name=Links" json:"Links,omitempty"`
	PageNum    int32    `protobuf:"varint,3,opt,name=PageNum" json:"PageNum,omitempty"`
	PageSize   int32    `protobuf:"varint,4,opt,name=PageSize" json:"PageSize,omitempty"`
	Count      bool     `protobuf:"varint,5,opt,name=Count" json:"Count,omitempty"`
}

func (m *Search) Reset()                    { *m = Search{} }
func (m *Search) String() string            { return proto.CompactTextString(m) }
func (*Search) ProtoMessage()               {}
func (*Search) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *Search) GetConditions() []string {
	if m != nil {
		return m.Conditions
	}
	return nil
}

func (m *Search) GetLinks() []string {
	if m != nil {
		return m.Links
	}
	return nil
}

func (m *Search) GetPageNum() int32 {
	if m != nil {
		return m.PageNum
	}
	return 0
}

func (m *Search) GetPageSize() int32 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

func (m *Search) GetCount() bool {
	if m != nil {
		return m.Count
	}
	return false
}

type Params struct {
	Obj    *google_protobuf1.Any `protobuf:"bytes,1,opt,name=Obj" json:"Obj,omitempty"`
	Ops    Ops                   `protobuf:"varint,2,opt,name=Ops,enum=message.Ops" json:"Ops,omitempty"`
	Search *Search               `protobuf:"bytes,3,opt,name=Search" json:"Search,omitempty"`
}

func (m *Params) Reset()                    { *m = Params{} }
func (m *Params) String() string            { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()               {}
func (*Params) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *Params) GetObj() *google_protobuf1.Any {
	if m != nil {
		return m.Obj
	}
	return nil
}

func (m *Params) GetOps() Ops {
	if m != nil {
		return m.Ops
	}
	return Ops_READ
}

func (m *Params) GetSearch() *Search {
	if m != nil {
		return m.Search
	}
	return nil
}

type Result struct {
	Succeed    bool                    `protobuf:"varint,1,opt,name=Succeed" json:"Succeed,omitempty"`
	MaxPageNum int32                   `protobuf:"varint,2,opt,name=MaxPageNum" json:"MaxPageNum,omitempty"`
	Objs       []*google_protobuf1.Any `protobuf:"bytes,3,rep,name=Objs" json:"Objs,omitempty"`
}

func (m *Result) Reset()                    { *m = Result{} }
func (m *Result) String() string            { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()               {}
func (*Result) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{2} }

func (m *Result) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *Result) GetMaxPageNum() int32 {
	if m != nil {
		return m.MaxPageNum
	}
	return 0
}

func (m *Result) GetObjs() []*google_protobuf1.Any {
	if m != nil {
		return m.Objs
	}
	return nil
}

type Job struct {
	// @inject_tag: form:"Name" gorm:"type:varchar(64);not null;unique_index:n_g_idx"
	Name string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty" form:"Name" gorm:"type:varchar(64);not null;unique_index:n_g_idx"`
	// @inject_tag: form:"Region" gorm:"type:varchar(64);not null;unique_index:n_g_idx"
	Region string `protobuf:"bytes,2,opt,name=Region" json:"Region,omitempty" form:"Region" gorm:"type:varchar(64);not null;unique_index:n_g_idx"`
	// @inject_tag: gorm:"type:varchar(64)"
	Schedule string `protobuf:"bytes,3,opt,name=Schedule" json:"Schedule,omitempty" gorm:"type:varchar(64)"`
	// @inject_tag: gorm:"type:tinyint(4)"
	Shell bool `protobuf:"varint,4,opt,name=Shell" json:"Shell,omitempty" gorm:"type:tinyint(4)"`
	// @inject_tag: gorm:"type:varchar(512);not null"
	Command string `protobuf:"bytes,5,opt,name=Command" json:"Command,omitempty" gorm:"type:varchar(512);not null"`
	// @inject_tag: gorm:"type:varchar(64);not null"
	Expression string `protobuf:"bytes,6,opt,name=Expression" json:"Expression,omitempty" gorm:"type:varchar(64);not null"`
	// @inject_tag: gorm:"type:tinyint(4)"
	Idempotent bool `protobuf:"varint,7,opt,name=Idempotent" json:"Idempotent,omitempty" gorm:"type:tinyint(4)"`
	// @inject_tag: gorm:"type:tinyint(4)"
	Disable bool `protobuf:"varint,8,opt,name=Disable" json:"Disable,omitempty" gorm:"type:tinyint(4)"`
	// @inject_tag: gorm:"type:varchar(64); not null"
	SchedulerNodeName string `protobuf:"bytes,9,opt,name=SchedulerNodeName" json:"SchedulerNodeName,omitempty" gorm:"type:varchar(64); not null"`
	// @inject_tag: gorm:"type:varchar(64);not null;primary_key"
	Id string `protobuf:"bytes,10,opt,name=Id" json:"Id,omitempty" gorm:"type:varchar(64);not null;primary_key"`
	// @inject_tag: gorm:"type:varchar(64)"
	ParentJob *Job `protobuf:"bytes,11,opt,name=ParentJob" json:"ParentJob,omitempty" gorm:"type:varchar(64)"`
}

func (m *Job) Reset()                    { *m = Job{} }
func (m *Job) String() string            { return proto.CompactTextString(m) }
func (*Job) ProtoMessage()               {}
func (*Job) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{3} }

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

func (m *Job) GetIdempotent() bool {
	if m != nil {
		return m.Idempotent
	}
	return false
}

func (m *Job) GetDisable() bool {
	if m != nil {
		return m.Disable
	}
	return false
}

func (m *Job) GetSchedulerNodeName() string {
	if m != nil {
		return m.SchedulerNodeName
	}
	return ""
}

func (m *Job) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Job) GetParentJob() *Job {
	if m != nil {
		return m.ParentJob
	}
	return nil
}

type JobStatus struct {
	// @inject_tag: form:"Name"
	Name string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty" form:"Name"`
	// @inject_tag: form:"Region"
	Region          string `protobuf:"bytes,2,opt,name=Region" json:"Region,omitempty" form:"Region"`
	SuccessCount    int64  `protobuf:"varint,3,opt,name=SuccessCount" json:"SuccessCount,omitempty"`
	ErrorCount      int64  `protobuf:"varint,4,opt,name=ErrorCount" json:"ErrorCount,omitempty"`
	LastHandleAgent string `protobuf:"bytes,5,opt,name=LastHandleAgent" json:"LastHandleAgent,omitempty"`
	LastSuccess     string `protobuf:"bytes,6,opt,name=LastSuccess" json:"LastSuccess,omitempty"`
	LastError       string `protobuf:"bytes,7,opt,name=LastError" json:"LastError,omitempty"`
}

func (m *JobStatus) Reset()                    { *m = JobStatus{} }
func (m *JobStatus) String() string            { return proto.CompactTextString(m) }
func (*JobStatus) ProtoMessage()               {}
func (*JobStatus) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{4} }

func (m *JobStatus) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *JobStatus) GetRegion() string {
	if m != nil {
		return m.Region
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
	// @inject_tag: gorm:"type:varchar(64);not null"
	SchedulerNodeName string `protobuf:"bytes,1,opt,name=SchedulerNodeName" json:"SchedulerNodeName,omitempty" gorm:"type:varchar(64);not null"`
	// @inject_tag: gorm:"type:blob"
	Output []byte `protobuf:"bytes,2,opt,name=Output,proto3" json:"Output,omitempty" gorm:"type:blob"`
	// @inject_tag: gorm:"type:tinyint(4)"
	Succeed bool `protobuf:"varint,3,opt,name=Succeed" json:"Succeed,omitempty" gorm:"type:tinyint(4)"`
	// @inject_tag: gorm:"type:bigint(21);not null"
	StartTime int64 `protobuf:"varint,4,opt,name=StartTime" json:"StartTime,omitempty" gorm:"type:bigint(21);not null"`
	// @inject_tag: gorm:"type:bigint(21);not null"
	FinishTime int64 `protobuf:"varint,5,opt,name=FinishTime" json:"FinishTime,omitempty" gorm:"type:bigint(21);not null"`
	// @inject_tag: gorm:"type:varchar(64);primary_key;not null" form:"Name"
	Name string `protobuf:"bytes,6,opt,name=Name" json:"Name,omitempty" gorm:"type:varchar(64);primary_key;not null" form:"Name"`
	// @inject_tag: gorm:"type:varchar(64);primary_key;not null" form:"Rgion"
	Region string `protobuf:"bytes,7,opt,name=Region" json:"Region,omitempty" gorm:"type:varchar(64);primary_key;not null" form:"Rgion"`
	// @inject_tag: gorm:"type:int(11);not null"
	Retries int64 `protobuf:"varint,8,opt,name=Retries" json:"Retries,omitempty" gorm:"type:int(11);not null"`
	// @inject_tag: gorm:"type:bigint(21);primary_key;not null" form:"Group"
	Group int64 `protobuf:"varint,9,opt,name=Group" json:"Group,omitempty" gorm:"type:bigint(21);primary_key;not null" form:"Group"`
	// @inject_tag: gorm:"type:varchar(64);primary_key;not null" form:"NodeName"
	RunNodeName string `protobuf:"bytes,10,opt,name=RunNodeName" json:"RunNodeName,omitempty" gorm:"type:varchar(64);primary_key;not null" form:"NodeName"`
}

func (m *Execution) Reset()                    { *m = Execution{} }
func (m *Execution) String() string            { return proto.CompactTextString(m) }
func (*Execution) ProtoMessage()               {}
func (*Execution) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{5} }

func (m *Execution) GetSchedulerNodeName() string {
	if m != nil {
		return m.SchedulerNodeName
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

func (m *Execution) GetStartTime() int64 {
	if m != nil {
		return m.StartTime
	}
	return 0
}

func (m *Execution) GetFinishTime() int64 {
	if m != nil {
		return m.FinishTime
	}
	return 0
}

func (m *Execution) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Execution) GetRegion() string {
	if m != nil {
		return m.Region
	}
	return ""
}

func (m *Execution) GetRetries() int64 {
	if m != nil {
		return m.Retries
	}
	return 0
}

func (m *Execution) GetGroup() int64 {
	if m != nil {
		return m.Group
	}
	return 0
}

func (m *Execution) GetRunNodeName() string {
	if m != nil {
		return m.RunNodeName
	}
	return ""
}

func init() {
	proto.RegisterType((*Search)(nil), "message.Search")
	proto.RegisterType((*Params)(nil), "message.Params")
	proto.RegisterType((*Result)(nil), "message.Result")
	proto.RegisterType((*Job)(nil), "message.Job")
	proto.RegisterType((*JobStatus)(nil), "message.JobStatus")
	proto.RegisterType((*Execution)(nil), "message.Execution")
	proto.RegisterEnum("message.Ops", Ops_name, Ops_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Job service

type JobClient interface {
	GetJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Job, error)
	ExecDone(ctx context.Context, in *Execution, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
	DoOps(ctx context.Context, in *Params, opts ...grpc.CallOption) (*Result, error)
	ProxyJobRun(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Execution, error)
}

type jobClient struct {
	cc *grpc.ClientConn
}

func NewJobClient(cc *grpc.ClientConn) JobClient {
	return &jobClient{cc}
}

func (c *jobClient) GetJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Job, error) {
	out := new(Job)
	err := grpc.Invoke(ctx, "/message.job/GetJob", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobClient) ExecDone(ctx context.Context, in *Execution, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/message.job/ExecDone", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobClient) DoOps(ctx context.Context, in *Params, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/message.job/DoOps", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobClient) ProxyJobRun(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Execution, error) {
	out := new(Execution)
	err := grpc.Invoke(ctx, "/message.job/ProxyJobRun", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Job service

type JobServer interface {
	GetJob(context.Context, *Job) (*Job, error)
	ExecDone(context.Context, *Execution) (*google_protobuf.Empty, error)
	DoOps(context.Context, *Params) (*Result, error)
	ProxyJobRun(context.Context, *Job) (*Execution, error)
}

func RegisterJobServer(s *grpc.Server, srv JobServer) {
	s.RegisterService(&_Job_serviceDesc, srv)
}

func _Job_GetJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Job)
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
		return srv.(JobServer).GetJob(ctx, req.(*Job))
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

func _Job_DoOps_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Params)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServer).DoOps(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message.job/DoOps",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServer).DoOps(ctx, req.(*Params))
	}
	return interceptor(ctx, in, info, handler)
}

func _Job_ProxyJobRun_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Job)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServer).ProxyJobRun(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message.job/ProxyJobRun",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServer).ProxyJobRun(ctx, req.(*Job))
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
			MethodName: "ExecDone",
			Handler:    _Job_ExecDone_Handler,
		},
		{
			MethodName: "DoOps",
			Handler:    _Job_DoOps_Handler,
		},
		{
			MethodName: "ProxyJobRun",
			Handler:    _Job_ProxyJobRun_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "message/job.proto",
}

func init() { proto.RegisterFile("message/job.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 753 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x54, 0xcf, 0x6e, 0xda, 0x4e,
	0x10, 0x8e, 0x31, 0x18, 0x3c, 0x44, 0x09, 0x59, 0x45, 0x91, 0x7f, 0xfc, 0xaa, 0x2a, 0xf2, 0xa1,
	0x45, 0x69, 0x45, 0x5a, 0x7a, 0xe9, 0x15, 0xc5, 0x24, 0x4d, 0x95, 0x84, 0x68, 0xc9, 0xa5, 0x47,
	0x03, 0x5b, 0x70, 0x8a, 0xbd, 0xc8, 0x7f, 0x24, 0xe8, 0x23, 0xf4, 0x45, 0xfa, 0x1c, 0x3d, 0xf5,
	0x49, 0xfa, 0x1e, 0x9d, 0x1d, 0xdb, 0xd8, 0xe4, 0xcf, 0xa1, 0x27, 0xf6, 0xfb, 0x66, 0xd9, 0xf9,
	0xe6, 0x9b, 0xf1, 0xc0, 0x81, 0x2f, 0xa2, 0xc8, 0x9d, 0x89, 0xd3, 0x7b, 0x39, 0xee, 0x2e, 0x43,
	0x19, 0x4b, 0x56, 0xcf, 0xa8, 0xf6, 0xff, 0x33, 0x29, 0x67, 0x0b, 0x71, 0x4a, 0xf4, 0x38, 0xf9,
	0x7a, 0x2a, 0xfc, 0x65, 0xbc, 0x4e, 0x6f, 0xb5, 0xff, 0x7b, 0x18, 0x74, 0x83, 0x2c, 0x64, 0xff,
	0xd0, 0xc0, 0x18, 0x09, 0x37, 0x9c, 0xcc, 0xd9, 0x4b, 0x80, 0x33, 0x19, 0x4c, 0xbd, 0xd8, 0x93,
	0x41, 0x64, 0x69, 0xc7, 0x7a, 0xc7, 0xe4, 0x25, 0x86, 0x1d, 0x42, 0xed, 0xca, 0x0b, 0xbe, 0x45,
	0x56, 0x85, 0x42, 0x29, 0x60, 0x16, 0xd4, 0x6f, 0x51, 0xc0, 0x4d, 0xe2, 0x5b, 0xfa, 0xb1, 0xd6,
	0xa9, 0xf1, 0x1c, 0xb2, 0x36, 0x34, 0xd4, 0x71, 0xe4, 0x7d, 0x17, 0x56, 0x95, 0x42, 0x1b, 0xac,
	0xde, 0x3a, 0x93, 0x49, 0x10, 0x5b, 0x35, 0x0c, 0x34, 0x78, 0x0a, 0xec, 0x35, 0x18, 0xb7, 0x6e,
	0xe8, 0xfa, 0x11, 0x7b, 0x05, 0xfa, 0x70, 0x7c, 0x8f, 0x22, 0xb4, 0x4e, 0xb3, 0x77, 0xd8, 0x4d,
	0xf5, 0x77, 0x73, 0xfd, 0xdd, 0x7e, 0xb0, 0xe6, 0xea, 0x02, 0x6a, 0xd6, 0x87, 0x4b, 0xa5, 0x48,
	0xeb, 0xec, 0xf5, 0x76, 0xbb, 0x99, 0x1b, 0x5d, 0xe4, 0xb8, 0x0a, 0xb0, 0xd7, 0x79, 0x75, 0x24,
	0xae, 0xd9, 0xdb, 0xdf, 0x5c, 0x49, 0x69, 0x9e, 0x85, 0xed, 0x05, 0x18, 0x5c, 0x44, 0xc9, 0x22,
	0x56, 0x05, 0x8d, 0x92, 0xc9, 0x44, 0x88, 0x29, 0xa5, 0x6f, 0xf0, 0x1c, 0x2a, 0x83, 0xae, 0xdd,
	0x55, 0x5e, 0x6d, 0x85, 0x4a, 0x2a, 0x31, 0xac, 0x03, 0x55, 0xd4, 0x14, 0x61, 0x2a, 0xfd, 0x59,
	0xd5, 0x74, 0xc3, 0xfe, 0x55, 0x01, 0xfd, 0xb3, 0x1c, 0x33, 0x06, 0xd5, 0x1b, 0xd7, 0x17, 0x94,
	0xc8, 0xe4, 0x74, 0x66, 0x47, 0x4a, 0xc9, 0x0c, 0x1d, 0xa7, 0x0c, 0x26, 0xcf, 0x90, 0xb2, 0x73,
	0x34, 0x99, 0x8b, 0x69, 0xb2, 0x10, 0x54, 0x8c, 0xc9, 0x37, 0x58, 0xd9, 0x39, 0x9a, 0x8b, 0xc5,
	0x82, 0x7c, 0x46, 0x3b, 0x09, 0xa8, 0x4a, 0xce, 0xa4, 0xef, 0xbb, 0xc1, 0x94, 0x6c, 0x36, 0x79,
	0x0e, 0x55, 0x25, 0x83, 0xd5, 0x32, 0x44, 0x2b, 0x54, 0x1e, 0x83, 0x82, 0x25, 0x46, 0xc5, 0x2f,
	0xa7, 0x38, 0x41, 0x32, 0x16, 0xd8, 0xa3, 0x3a, 0x3d, 0x5a, 0x62, 0xd4, 0xcb, 0x8e, 0x17, 0xb9,
	0x63, 0x94, 0xd2, 0x48, 0x3d, 0xca, 0x20, 0x7b, 0x0b, 0x07, 0xb9, 0xaa, 0xf0, 0x46, 0x4e, 0x05,
	0x95, 0x67, 0x52, 0x82, 0xc7, 0x01, 0xb6, 0x07, 0x95, 0xcb, 0xa9, 0x05, 0x14, 0xc6, 0x13, 0x3b,
	0x01, 0x13, 0x07, 0x00, 0x33, 0xa0, 0x39, 0x56, 0x93, 0x3a, 0x56, 0x34, 0x15, 0x39, 0x5e, 0x84,
	0xed, 0x3f, 0x1a, 0x98, 0xf8, 0x3b, 0x8a, 0xdd, 0x38, 0x89, 0xfe, 0xc9, 0x49, 0x1b, 0x76, 0xa9,
	0xa5, 0x51, 0x94, 0xce, 0xa0, 0x72, 0x53, 0xe7, 0x5b, 0x1c, 0x39, 0x14, 0x86, 0x32, 0x4c, 0x6f,
	0x54, 0xe9, 0x46, 0x89, 0xc1, 0x5e, 0xef, 0x5f, 0xb9, 0x51, 0xfc, 0x09, 0xdd, 0x5c, 0x88, 0xfe,
	0x4c, 0x64, 0xa3, 0x6c, 0xf2, 0x87, 0x34, 0x3b, 0x86, 0xa6, 0xa2, 0xb2, 0xd7, 0x33, 0xb3, 0xcb,
	0x14, 0x7b, 0x01, 0xa6, 0x82, 0xf4, 0x3a, 0x99, 0x6d, 0xf2, 0x82, 0xb0, 0x7f, 0x56, 0xc0, 0x1c,
	0xac, 0xc4, 0x24, 0x51, 0x5f, 0xe1, 0xd3, 0xfe, 0x6a, 0xcf, 0xf9, 0x8b, 0x0e, 0x0c, 0x93, 0x78,
	0x99, 0xc4, 0xe4, 0xc0, 0x2e, 0xcf, 0x50, 0x79, 0xc6, 0xf5, 0xed, 0x19, 0x47, 0x2d, 0xe8, 0x68,
	0x18, 0xdf, 0x79, 0xbe, 0xc8, 0xca, 0x2e, 0x08, 0xe5, 0xca, 0xb9, 0x17, 0x78, 0xd1, 0x9c, 0xc2,
	0xb5, 0xd4, 0x95, 0x82, 0xd9, 0x74, 0xc1, 0x78, 0xb2, 0x0b, 0xf5, 0xad, 0x2e, 0xa0, 0x06, 0x2e,
	0xe2, 0xd0, 0x13, 0x11, 0xcd, 0x90, 0xce, 0x73, 0xa8, 0xa6, 0xf9, 0x22, 0x94, 0xc9, 0x92, 0xe6,
	0x46, 0xe7, 0x29, 0x50, 0x3e, 0xf2, 0x24, 0xd8, 0xd4, 0x9c, 0x0e, 0x4d, 0x99, 0x3a, 0x79, 0x47,
	0xcb, 0x80, 0x35, 0xa0, 0xca, 0x07, 0x7d, 0xa7, 0xb5, 0xc3, 0xea, 0xa0, 0xf7, 0x1d, 0xa7, 0xa5,
	0x31, 0x00, 0xe3, 0x7a, 0xe8, 0x5c, 0x9e, 0x7f, 0x69, 0x55, 0xd4, 0xd9, 0x19, 0x5c, 0x0d, 0xee,
	0x06, 0x2d, 0xbd, 0xf7, 0x5b, 0x03, 0x1d, 0x97, 0x29, 0xae, 0x1b, 0xe3, 0x42, 0xa8, 0xa9, 0x62,
	0x5b, 0xe3, 0xd6, 0xde, 0x42, 0xf6, 0x0e, 0xfb, 0x08, 0x0d, 0xd5, 0x0a, 0x47, 0x06, 0x58, 0xeb,
	0x26, 0xb6, 0xe9, 0x4e, 0xfb, 0xe8, 0xd1, 0x37, 0x3f, 0x50, 0x6b, 0x18, 0xff, 0xf9, 0x06, 0x6a,
	0x8e, 0x54, 0xea, 0x8a, 0x0d, 0x94, 0xae, 0xba, 0x76, 0x41, 0xa4, 0x0b, 0x08, 0x2f, 0xbf, 0x87,
	0xe6, 0x6d, 0x28, 0x57, 0x6b, 0x35, 0xf1, 0x49, 0xf0, 0x40, 0xd3, 0x13, 0x79, 0xed, 0x9d, 0xb1,
	0x41, 0x19, 0x3f, 0xfc, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x45, 0x6b, 0x27, 0x53, 0x24, 0x06, 0x00,
	0x00,
}
