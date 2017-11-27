// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message/job.proto

package message

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/empty"
import google_protobuf1 "github.com/golang/protobuf/ptypes/any"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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
	// @inject_tag: gorm:"type:varchar(64);not null;primary_key"
	Name string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty" gorm:"type:varchar(64);not null;primary_key"`
	// @inject_tag: gorm:"type:varchar(64);not null;primary_key"
	Region string `protobuf:"bytes,2,opt,name=Region" json:"Region,omitempty" gorm:"type:varchar(64);not null;primary_key"`
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
	// @inject_tag: gorm:"type:varchar(64);not null;index:pjn_idx"
	ParentJobName string `protobuf:"bytes,10,opt,name=ParentJobName" json:"ParentJobName,omitempty" gorm:"type:varchar(64);not null;index:pjn_idx"`
	// @inject_tag: gorm:"-"
	ParentJob *Job `protobuf:"bytes,11,opt,name=ParentJob" json:"ParentJob,omitempty" gorm:"-"`
	// @inject_tag: gorm:"type:decimal(3,2);not null;default:'1.00'"
	Threshold float64 `protobuf:"fixed64,12,opt,name=Threshold" json:"Threshold,omitempty" gorm:"type:decimal(3,2);not null;default:'1.00'"`
	// @inject_tag: gorm:"type:bigint(21);not null;default:'3600'"
	MaxRunTime int32 `protobuf:"varint,13,opt,name=MaxRunTime" json:"MaxRunTime,omitempty" gorm:"type:bigint(21);not null;default:'3600'"`
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

func (m *Job) GetParentJobName() string {
	if m != nil {
		return m.ParentJobName
	}
	return ""
}

func (m *Job) GetParentJob() *Job {
	if m != nil {
		return m.ParentJob
	}
	return nil
}

func (m *Job) GetThreshold() float64 {
	if m != nil {
		return m.Threshold
	}
	return 0
}

func (m *Job) GetMaxRunTime() int32 {
	if m != nil {
		return m.MaxRunTime
	}
	return 0
}

type JobStatus struct {
	Name            string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	Region          string `protobuf:"bytes,2,opt,name=Region" json:"Region,omitempty"`
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

type Node struct {
	Name    string            `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	Region  string            `protobuf:"bytes,2,opt,name=Region" json:"Region,omitempty"`
	Version string            `protobuf:"bytes,3,opt,name=Version" json:"Version,omitempty"`
	Alived  bool              `protobuf:"varint,4,opt,name=Alived" json:"Alived,omitempty"`
	Role    string            `protobuf:"bytes,5,opt,name=Role" json:"Role,omitempty"`
	Tags    map[string]string `protobuf:"bytes,6,rep,name=Tags" json:"Tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Node) Reset()                    { *m = Node{} }
func (m *Node) String() string            { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()               {}
func (*Node) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{6} }

func (m *Node) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Node) GetRegion() string {
	if m != nil {
		return m.Region
	}
	return ""
}

func (m *Node) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *Node) GetAlived() bool {
	if m != nil {
		return m.Alived
	}
	return false
}

func (m *Node) GetRole() string {
	if m != nil {
		return m.Role
	}
	return ""
}

func (m *Node) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func init() {
	proto.RegisterType((*Search)(nil), "message.Search")
	proto.RegisterType((*Params)(nil), "message.Params")
	proto.RegisterType((*Result)(nil), "message.Result")
	proto.RegisterType((*Job)(nil), "message.Job")
	proto.RegisterType((*JobStatus)(nil), "message.JobStatus")
	proto.RegisterType((*Execution)(nil), "message.Execution")
	proto.RegisterType((*Node)(nil), "message.Node")
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
	// 886 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x55, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0x16, 0x45, 0x89, 0x12, 0x47, 0x4e, 0xa3, 0x2c, 0x02, 0x97, 0x55, 0x83, 0x40, 0x20, 0x8a,
	0x54, 0x48, 0x0a, 0xb9, 0x75, 0x0f, 0x0d, 0x7a, 0x13, 0x2c, 0x25, 0x8d, 0xe1, 0x58, 0xc6, 0xca,
	0x28, 0xd0, 0x23, 0x29, 0x4e, 0x25, 0x3a, 0x24, 0x57, 0xe0, 0x92, 0x81, 0xd5, 0x47, 0xe8, 0xbd,
	0xcf, 0xd0, 0x47, 0xe9, 0x5b, 0xf4, 0xd6, 0xf7, 0x28, 0x66, 0xf8, 0x23, 0xc9, 0x71, 0x0e, 0xbe,
	0xed, 0xf7, 0xcd, 0x70, 0xe7, 0xe7, 0x9b, 0x59, 0xc2, 0x93, 0x18, 0xb5, 0xf6, 0x56, 0x78, 0x72,
	0xa3, 0xfc, 0xf1, 0x26, 0x55, 0x99, 0x12, 0x9d, 0x92, 0x1a, 0x7c, 0xbd, 0x52, 0x6a, 0x15, 0xe1,
	0x09, 0xd3, 0x7e, 0xfe, 0xfb, 0x09, 0xc6, 0x9b, 0x6c, 0x5b, 0x78, 0x0d, 0xbe, 0xba, 0x6b, 0xf4,
	0x92, 0xd2, 0xe4, 0xfe, 0x69, 0x80, 0xb5, 0x40, 0x2f, 0x5d, 0xae, 0xc5, 0x73, 0x80, 0x33, 0x95,
	0x04, 0x61, 0x16, 0xaa, 0x44, 0x3b, 0xc6, 0xd0, 0x1c, 0xd9, 0x72, 0x8f, 0x11, 0x4f, 0xa1, 0x7d,
	0x11, 0x26, 0x1f, 0xb4, 0xd3, 0x64, 0x53, 0x01, 0x84, 0x03, 0x9d, 0x2b, 0x6f, 0x85, 0x97, 0x79,
	0xec, 0x98, 0x43, 0x63, 0xd4, 0x96, 0x15, 0x14, 0x03, 0xe8, 0xd2, 0x71, 0x11, 0xfe, 0x81, 0x4e,
	0x8b, 0x4d, 0x35, 0xa6, 0xbb, 0xce, 0x54, 0x9e, 0x64, 0x4e, 0x7b, 0x68, 0x8c, 0xba, 0xb2, 0x00,
	0xee, 0x16, 0xac, 0x2b, 0x2f, 0xf5, 0x62, 0x2d, 0x5e, 0x80, 0x39, 0xf7, 0x6f, 0x1c, 0x63, 0x68,
	0x8c, 0x7a, 0xa7, 0x4f, 0xc7, 0x45, 0xfe, 0xe3, 0x2a, 0xff, 0xf1, 0x24, 0xd9, 0x4a, 0x72, 0x10,
	0xcf, 0xc1, 0x9c, 0x6f, 0x28, 0x23, 0x63, 0xf4, 0xc5, 0xe9, 0xd1, 0xb8, 0xec, 0xc6, 0x78, 0xbe,
	0xd1, 0x92, 0x0c, 0xe2, 0xdb, 0xaa, 0x3a, 0x4e, 0xae, 0x77, 0xfa, 0xb8, 0x76, 0x29, 0x68, 0x59,
	0x9a, 0xdd, 0x08, 0x2c, 0x89, 0x3a, 0x8f, 0x32, 0x2a, 0x68, 0x91, 0x2f, 0x97, 0x88, 0x01, 0x87,
	0xef, 0xca, 0x0a, 0x52, 0x83, 0xde, 0x7b, 0xb7, 0x55, 0xb5, 0x4d, 0x2e, 0x69, 0x8f, 0x11, 0x23,
	0x68, 0xcd, 0xfd, 0x1b, 0xed, 0x98, 0x43, 0xf3, 0xb3, 0x59, 0xb3, 0x87, 0xfb, 0x97, 0x09, 0xe6,
	0xb9, 0xf2, 0x85, 0x80, 0xd6, 0xa5, 0x17, 0x23, 0x07, 0xb2, 0x25, 0x9f, 0xc5, 0x31, 0x65, 0xb2,
	0x0a, 0x55, 0xc2, 0x11, 0x6c, 0x59, 0x22, 0x6a, 0xe7, 0x62, 0xb9, 0xc6, 0x20, 0x8f, 0x90, 0x8b,
	0xb1, 0x65, 0x8d, 0xa9, 0x9d, 0x8b, 0x35, 0x46, 0x11, 0xf7, 0xb9, 0x2b, 0x0b, 0x40, 0x95, 0x9c,
	0xa9, 0x38, 0xf6, 0x92, 0x80, 0xdb, 0x6c, 0xcb, 0x0a, 0x52, 0x25, 0xb3, 0xdb, 0x4d, 0x8a, 0x5a,
	0x53, 0x1c, 0x8b, 0x8d, 0x7b, 0x0c, 0xd9, 0xdf, 0x05, 0x18, 0x6f, 0x54, 0x86, 0x49, 0xe6, 0x74,
	0xf8, 0xd2, 0x3d, 0x86, 0x6e, 0x9e, 0x86, 0xda, 0xf3, 0x23, 0x74, 0xba, 0x45, 0x8f, 0x4a, 0x28,
	0xbe, 0x83, 0x27, 0x55, 0x56, 0xe9, 0xa5, 0x0a, 0x90, 0xcb, 0xb3, 0x39, 0xc0, 0xa7, 0x06, 0xf1,
	0x0d, 0x3c, 0xba, 0xf2, 0x52, 0x4c, 0xb2, 0x73, 0xe5, 0xb3, 0x27, 0xb0, 0xe7, 0x21, 0x29, 0x5e,
	0x82, 0x5d, 0x13, 0x4e, 0x8f, 0x75, 0xdc, 0x49, 0x7d, 0xae, 0x7c, 0xb9, 0x33, 0x8b, 0x67, 0x60,
	0x5f, 0xaf, 0x53, 0xd4, 0x6b, 0x15, 0x05, 0xce, 0xd1, 0xd0, 0x18, 0x19, 0x72, 0x47, 0x94, 0x0a,
	0xca, 0x3c, 0xb9, 0x0e, 0x63, 0x74, 0x1e, 0xd5, 0x0a, 0x96, 0x8c, 0xfb, 0x9f, 0x01, 0xf6, 0xb9,
	0xf2, 0x17, 0x99, 0x97, 0xe5, 0xfa, 0x41, 0xea, 0xb8, 0x70, 0xc4, 0x63, 0xa2, 0x75, 0x31, 0xd7,
	0xa4, 0x90, 0x29, 0x0f, 0x38, 0xee, 0x7a, 0x9a, 0xaa, 0xb4, 0xf0, 0x68, 0xb1, 0xc7, 0x1e, 0x23,
	0x46, 0xf0, 0xf8, 0xc2, 0xd3, 0xd9, 0x2f, 0x5e, 0x12, 0x44, 0x38, 0x59, 0x61, 0xb9, 0x1e, 0xb6,
	0xbc, 0x4b, 0x8b, 0x21, 0xf4, 0x88, 0x2a, 0x6f, 0x2f, 0x05, 0xdc, 0xa7, 0xa8, 0x0f, 0x04, 0xf9,
	0x76, 0x16, 0xd0, 0x96, 0x3b, 0xc2, 0xfd, 0xbb, 0x09, 0xf6, 0xec, 0x16, 0x97, 0x39, 0x6d, 0xf6,
	0xfd, 0x9a, 0x19, 0x9f, 0xd3, 0xec, 0x18, 0xac, 0x79, 0x9e, 0x6d, 0xf2, 0x8c, 0x3b, 0x70, 0x24,
	0x4b, 0xb4, 0xbf, 0x37, 0xe6, 0xe1, 0xde, 0x3c, 0x03, 0x7b, 0x91, 0x79, 0x69, 0xc6, 0x4d, 0x2f,
	0xca, 0xde, 0x11, 0xd4, 0x95, 0x37, 0x61, 0x12, 0xea, 0x35, 0x9b, 0xdb, 0x45, 0x57, 0x76, 0x4c,
	0xad, 0x82, 0x75, 0xaf, 0x0a, 0x9d, 0x03, 0x15, 0x1c, 0xe8, 0x48, 0xcc, 0xd2, 0x10, 0x35, 0xcf,
	0xa5, 0x29, 0x2b, 0x48, 0x1b, 0xf2, 0x36, 0x55, 0xf9, 0x86, 0x67, 0xd1, 0x94, 0x05, 0xa0, 0x3e,
	0xca, 0x3c, 0xa9, 0x6b, 0x2e, 0xa6, 0x6f, 0x9f, 0x72, 0xff, 0x35, 0xa0, 0x45, 0xe0, 0x41, 0xc3,
	0xe0, 0x40, 0xe7, 0x57, 0x4c, 0x79, 0xb7, 0x8a, 0x4d, 0xad, 0x20, 0x7d, 0x31, 0x89, 0xc2, 0x8f,
	0x18, 0x94, 0x9b, 0x5a, 0x22, 0xba, 0x5d, 0xaa, 0x08, 0x4b, 0xbd, 0xf9, 0x2c, 0x5e, 0x41, 0xeb,
	0xda, 0x5b, 0x91, 0xba, 0xf4, 0x9c, 0x7c, 0x59, 0x4f, 0x3c, 0xa5, 0x33, 0x26, 0xcb, 0x2c, 0xc9,
	0xd2, 0xad, 0x64, 0xa7, 0xc1, 0x4f, 0x60, 0xd7, 0x94, 0xe8, 0x83, 0xf9, 0x01, 0xb7, 0x65, 0xaa,
	0x74, 0xa4, 0xf2, 0x3f, 0x7a, 0x51, 0x8e, 0x65, 0xa2, 0x05, 0xf8, 0xb9, 0xf9, 0xda, 0x78, 0xf9,
	0x3d, 0xbf, 0xa0, 0xa2, 0x0b, 0x2d, 0x39, 0x9b, 0x4c, 0xfb, 0x0d, 0xd1, 0x01, 0x73, 0x32, 0x9d,
	0xf6, 0x0d, 0x01, 0x60, 0xbd, 0x9f, 0x4f, 0xdf, 0xbd, 0xf9, 0xad, 0xdf, 0xa4, 0xf3, 0x74, 0x76,
	0x31, 0xbb, 0x9e, 0xf5, 0xcd, 0xd3, 0x7f, 0x0c, 0x30, 0x6f, 0x94, 0x2f, 0x5e, 0x80, 0xf5, 0x16,
	0x79, 0xe9, 0x0e, 0xb6, 0x71, 0x70, 0x80, 0xdc, 0x86, 0x78, 0x0d, 0x5d, 0x9a, 0xb5, 0xa9, 0x4a,
	0x50, 0x88, 0xda, 0x56, 0x8f, 0xdf, 0xe0, 0xf8, 0x93, 0x87, 0x72, 0x46, 0xff, 0x2e, 0xb7, 0x21,
	0x5e, 0x41, 0x7b, 0xaa, 0x28, 0xbb, 0xdd, 0xb3, 0x5d, 0xfc, 0x1f, 0x06, 0x3b, 0xa2, 0x78, 0xb5,
	0xdd, 0x86, 0xf8, 0x01, 0x7a, 0x57, 0xa9, 0xba, 0xdd, 0xd2, 0x83, 0x90, 0x27, 0x77, 0x72, 0xba,
	0x27, 0xae, 0xdb, 0xf0, 0x2d, 0x8e, 0xf8, 0xe3, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0x12, 0xab,
	0xa0, 0xa4, 0x59, 0x07, 0x00, 0x00,
}
