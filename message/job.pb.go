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

type RequestJobAndToken struct {
	NodeName       string `protobuf:"bytes,1,opt,name=NodeName" json:"NodeName,omitempty"`
	JobName        string `protobuf:"bytes,2,opt,name=JobName" json:"JobName,omitempty"`
	JobRegion      string `protobuf:"bytes,3,opt,name=JobRegion" json:"JobRegion,omitempty"`
	ExecutionGroup int64  `protobuf:"varint,4,opt,name=ExecutionGroup" json:"ExecutionGroup,omitempty"`
}

func (m *RequestJobAndToken) Reset()                    { *m = RequestJobAndToken{} }
func (m *RequestJobAndToken) String() string            { return proto.CompactTextString(m) }
func (*RequestJobAndToken) ProtoMessage()               {}
func (*RequestJobAndToken) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *RequestJobAndToken) GetNodeName() string {
	if m != nil {
		return m.NodeName
	}
	return ""
}

func (m *RequestJobAndToken) GetJobName() string {
	if m != nil {
		return m.JobName
	}
	return ""
}

func (m *RequestJobAndToken) GetJobRegion() string {
	if m != nil {
		return m.JobRegion
	}
	return ""
}

func (m *RequestJobAndToken) GetExecutionGroup() int64 {
	if m != nil {
		return m.ExecutionGroup
	}
	return 0
}

type ResponseJobAndToken struct {
	Succeed  bool  `protobuf:"varint,1,opt,name=Succeed" json:"Succeed,omitempty"`
	Duration int64 `protobuf:"varint,2,opt,name=Duration" json:"Duration,omitempty"`
	Job      *Job  `protobuf:"bytes,3,opt,name=Job" json:"Job,omitempty"`
}

func (m *ResponseJobAndToken) Reset()                    { *m = ResponseJobAndToken{} }
func (m *ResponseJobAndToken) String() string            { return proto.CompactTextString(m) }
func (*ResponseJobAndToken) ProtoMessage()               {}
func (*ResponseJobAndToken) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *ResponseJobAndToken) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *ResponseJobAndToken) GetDuration() int64 {
	if m != nil {
		return m.Duration
	}
	return 0
}

func (m *ResponseJobAndToken) GetJob() *Job {
	if m != nil {
		return m.Job
	}
	return nil
}

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
func (*Search) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{2} }

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
func (*Params) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{3} }

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
func (*Result) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{4} }

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
	// @inject_tag: gorm:"type:int;not null;default:'-1'"
	Concurrency int32 `protobuf:"varint,14,opt,name=Concurrency" json:"Concurrency,omitempty" gorm:"type:int;not null;default:'-1'"`
}

func (m *Job) Reset()                    { *m = Job{} }
func (m *Job) String() string            { return proto.CompactTextString(m) }
func (*Job) ProtoMessage()               {}
func (*Job) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{5} }

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

func (m *Job) GetConcurrency() int32 {
	if m != nil {
		return m.Concurrency
	}
	return 0
}

type JobStatus struct {
	Name            string                   `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	Region          string                   `protobuf:"bytes,2,opt,name=Region" json:"Region,omitempty"`
	SuccessCount    int64                    `protobuf:"varint,3,opt,name=SuccessCount" json:"SuccessCount,omitempty"`
	ErrorCount      int64                    `protobuf:"varint,4,opt,name=ErrorCount" json:"ErrorCount,omitempty"`
	LastHandleAgent string                   `protobuf:"bytes,5,opt,name=LastHandleAgent" json:"LastHandleAgent,omitempty"`
	LastSuccess     string                   `protobuf:"bytes,6,opt,name=LastSuccess" json:"LastSuccess,omitempty"`
	LastError       string                   `protobuf:"bytes,7,opt,name=LastError" json:"LastError,omitempty"`
	RunningStatus   map[int64]*RunningStatus `protobuf:"bytes,8,rep,name=RunningStatus" json:"RunningStatus,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *JobStatus) Reset()                    { *m = JobStatus{} }
func (m *JobStatus) String() string            { return proto.CompactTextString(m) }
func (*JobStatus) ProtoMessage()               {}
func (*JobStatus) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{6} }

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

func (m *JobStatus) GetRunningStatus() map[int64]*RunningStatus {
	if m != nil {
		return m.RunningStatus
	}
	return nil
}

type RunningStatus struct {
	LeftToken   int32    `protobuf:"varint,1,opt,name=LeftToken" json:"LeftToken,omitempty"`
	RunningNode []string `protobuf:"bytes,2,rep,name=RunningNode" json:"RunningNode,omitempty"`
}

func (m *RunningStatus) Reset()                    { *m = RunningStatus{} }
func (m *RunningStatus) String() string            { return proto.CompactTextString(m) }
func (*RunningStatus) ProtoMessage()               {}
func (*RunningStatus) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{7} }

func (m *RunningStatus) GetLeftToken() int32 {
	if m != nil {
		return m.LeftToken
	}
	return 0
}

func (m *RunningStatus) GetRunningNode() []string {
	if m != nil {
		return m.RunningNode
	}
	return nil
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
func (*Execution) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{8} }

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
func (*Node) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{9} }

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
	proto.RegisterType((*RequestJobAndToken)(nil), "message.RequestJobAndToken")
	proto.RegisterType((*ResponseJobAndToken)(nil), "message.ResponseJobAndToken")
	proto.RegisterType((*Search)(nil), "message.Search")
	proto.RegisterType((*Params)(nil), "message.Params")
	proto.RegisterType((*Result)(nil), "message.Result")
	proto.RegisterType((*Job)(nil), "message.Job")
	proto.RegisterType((*JobStatus)(nil), "message.JobStatus")
	proto.RegisterType((*RunningStatus)(nil), "message.RunningStatus")
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
	GetJobAndToken(ctx context.Context, in *RequestJobAndToken, opts ...grpc.CallOption) (*ResponseJobAndToken, error)
	SendBackExecutionAndToken(ctx context.Context, in *Execution, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
	DoOps(ctx context.Context, in *Params, opts ...grpc.CallOption) (*Result, error)
	ProxyJobRun(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Execution, error)
}

type jobClient struct {
	cc *grpc.ClientConn
}

func NewJobClient(cc *grpc.ClientConn) JobClient {
	return &jobClient{cc}
}

func (c *jobClient) GetJobAndToken(ctx context.Context, in *RequestJobAndToken, opts ...grpc.CallOption) (*ResponseJobAndToken, error) {
	out := new(ResponseJobAndToken)
	err := grpc.Invoke(ctx, "/message.job/GetJobAndToken", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobClient) SendBackExecutionAndToken(ctx context.Context, in *Execution, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/message.job/SendBackExecutionAndToken", in, out, c.cc, opts...)
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
	GetJobAndToken(context.Context, *RequestJobAndToken) (*ResponseJobAndToken, error)
	SendBackExecutionAndToken(context.Context, *Execution) (*google_protobuf.Empty, error)
	DoOps(context.Context, *Params) (*Result, error)
	ProxyJobRun(context.Context, *Job) (*Execution, error)
}

func RegisterJobServer(s *grpc.Server, srv JobServer) {
	s.RegisterService(&_Job_serviceDesc, srv)
}

func _Job_GetJobAndToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestJobAndToken)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServer).GetJobAndToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message.job/GetJobAndToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServer).GetJobAndToken(ctx, req.(*RequestJobAndToken))
	}
	return interceptor(ctx, in, info, handler)
}

func _Job_SendBackExecutionAndToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Execution)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServer).SendBackExecutionAndToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message.job/SendBackExecutionAndToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServer).SendBackExecutionAndToken(ctx, req.(*Execution))
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
			MethodName: "GetJobAndToken",
			Handler:    _Job_GetJobAndToken_Handler,
		},
		{
			MethodName: "SendBackExecutionAndToken",
			Handler:    _Job_SendBackExecutionAndToken_Handler,
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
	// 1077 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x56, 0xdd, 0x6e, 0xe3, 0x44,
	0x14, 0xae, 0xe3, 0xfc, 0xf9, 0xa4, 0xdb, 0x76, 0x87, 0xaa, 0x78, 0xb3, 0x55, 0x15, 0x59, 0xb0,
	0x44, 0xbb, 0xab, 0x14, 0xca, 0x05, 0x88, 0xbb, 0xd0, 0x64, 0x97, 0x96, 0xb6, 0xa9, 0x26, 0x15,
	0x82, 0x4b, 0x3b, 0x9e, 0x4d, 0xdc, 0x3a, 0x33, 0xc1, 0x63, 0xaf, 0x1a, 0x1e, 0x01, 0x71, 0xc1,
	0x5b, 0xf0, 0x08, 0xbc, 0x0d, 0xaf, 0x02, 0x9a, 0x33, 0x63, 0xc7, 0x69, 0xba, 0x48, 0x7b, 0xe7,
	0xef, 0x3b, 0x67, 0x66, 0xce, 0xcf, 0x37, 0x67, 0x0c, 0x4f, 0xe7, 0x4c, 0x4a, 0x7f, 0xca, 0x8e,
	0x6f, 0x45, 0xd0, 0x5b, 0x24, 0x22, 0x15, 0xa4, 0x61, 0xa8, 0xf6, 0xf3, 0xa9, 0x10, 0xd3, 0x98,
	0x1d, 0x23, 0x1d, 0x64, 0xef, 0x8e, 0xd9, 0x7c, 0x91, 0x2e, 0xb5, 0x57, 0xfb, 0xd9, 0x43, 0xa3,
	0xcf, 0x8d, 0xc9, 0xfb, 0xd3, 0x02, 0x42, 0xd9, 0xaf, 0x19, 0x93, 0xe9, 0xb9, 0x08, 0xfa, 0x3c,
	0xbc, 0x11, 0x77, 0x8c, 0x93, 0x36, 0x34, 0xaf, 0x44, 0xc8, 0xae, 0xfc, 0x39, 0x73, 0xad, 0x8e,
	0xd5, 0x75, 0x68, 0x81, 0x89, 0x0b, 0x8d, 0x73, 0x11, 0xa0, 0xa9, 0x82, 0xa6, 0x1c, 0x92, 0x43,
	0x70, 0xce, 0x45, 0x40, 0xd9, 0x34, 0x12, 0xdc, 0xb5, 0xd1, 0xb6, 0x22, 0xc8, 0x0b, 0xd8, 0x19,
	0xde, 0xb3, 0x49, 0x96, 0x46, 0x82, 0xbf, 0x4d, 0x44, 0xb6, 0x70, 0xab, 0x1d, 0xab, 0x6b, 0xd3,
	0x07, 0xac, 0x77, 0x07, 0x9f, 0x50, 0x26, 0x17, 0x82, 0x4b, 0x56, 0x0e, 0xc9, 0x85, 0xc6, 0x38,
	0x9b, 0x4c, 0x18, 0x0b, 0x31, 0xa2, 0x26, 0xcd, 0xa1, 0x0a, 0x76, 0x90, 0x25, 0xbe, 0xda, 0x01,
	0x23, 0xb2, 0x69, 0x81, 0xc9, 0x11, 0xd8, 0xe7, 0x22, 0xc0, 0x60, 0x5a, 0x27, 0xdb, 0x3d, 0x53,
	0xae, 0x9e, 0x8a, 0x4a, 0x19, 0xbc, 0xdf, 0x2d, 0xa8, 0x8f, 0x99, 0x9f, 0x4c, 0x66, 0xe4, 0x08,
	0xe0, 0x54, 0xf0, 0x30, 0x52, 0xeb, 0xa4, 0x6b, 0x75, 0xec, 0xae, 0x43, 0x4b, 0x0c, 0xd9, 0x87,
	0xda, 0x45, 0xc4, 0xef, 0xa4, 0x5b, 0x41, 0x93, 0x06, 0x2a, 0xac, 0x6b, 0x7f, 0xca, 0xae, 0xb2,
	0x39, 0x1e, 0x52, 0xa3, 0x39, 0x54, 0x61, 0xa9, 0xcf, 0x71, 0xf4, 0x1b, 0xc3, 0x4c, 0x6b, 0xb4,
	0xc0, 0x6a, 0xaf, 0x53, 0x91, 0xf1, 0xd4, 0xad, 0x61, 0x2a, 0x1a, 0x78, 0x4b, 0xa8, 0x5f, 0xfb,
	0x89, 0x3f, 0x97, 0xe4, 0x05, 0xd8, 0xa3, 0xe0, 0x16, 0x13, 0x6d, 0x9d, 0xec, 0xf7, 0x74, 0xff,
	0x7a, 0x79, 0xff, 0x7a, 0x7d, 0xbe, 0xa4, 0xca, 0x41, 0xa5, 0x37, 0x5a, 0x48, 0xcc, 0x7a, 0xa7,
	0x94, 0xde, 0x68, 0x21, 0xa9, 0x32, 0x90, 0x2f, 0xf2, 0xec, 0x4c, 0x05, 0x76, 0x0b, 0x17, 0x4d,
	0x53, 0x63, 0xf6, 0x62, 0xa8, 0x53, 0x26, 0xb3, 0x38, 0xfd, 0x9f, 0x3a, 0x1f, 0x01, 0x5c, 0xfa,
	0xf7, 0x79, 0xb6, 0x15, 0x4c, 0xa9, 0xc4, 0x90, 0x2e, 0x54, 0x47, 0xc1, 0xad, 0x74, 0xed, 0x8e,
	0xfd, 0xc1, 0xa8, 0xd1, 0xc3, 0xfb, 0xdb, 0xc6, 0xb6, 0x10, 0x02, 0xd5, 0x92, 0xc4, 0xf0, 0x9b,
	0x1c, 0xa8, 0x48, 0xa6, 0x79, 0x2f, 0x1d, 0x6a, 0x90, 0x2a, 0xe7, 0x78, 0x32, 0x63, 0x61, 0x16,
	0x33, 0xa3, 0xad, 0x02, 0xab, 0x72, 0x8e, 0x67, 0x2c, 0x8e, 0xb1, 0xce, 0x4d, 0xaa, 0x81, 0xca,
	0xe4, 0x54, 0xcc, 0xe7, 0x3e, 0x0f, 0xb1, 0xcc, 0x0e, 0xcd, 0xa1, 0xca, 0x64, 0x78, 0xbf, 0x48,
	0x98, 0x94, 0xea, 0x9c, 0x3a, 0x1a, 0x4b, 0x8c, 0xb2, 0x9f, 0x85, 0x6c, 0xbe, 0x10, 0x29, 0xe3,
	0xa9, 0xdb, 0xc0, 0x4d, 0x4b, 0x8c, 0xda, 0x79, 0x10, 0x49, 0x3f, 0x88, 0x99, 0xdb, 0xd4, 0x35,
	0x32, 0x90, 0xbc, 0x86, 0xa7, 0x79, 0x54, 0x49, 0x71, 0x83, 0x1c, 0x3c, 0x60, 0xd3, 0x40, 0x3e,
	0x83, 0x27, 0xd7, 0x7e, 0xc2, 0x78, 0x9a, 0x5f, 0x28, 0x40, 0xcf, 0x75, 0x92, 0xbc, 0x04, 0xa7,
	0x20, 0xdc, 0xd6, 0x23, 0x4a, 0x5e, 0x99, 0xd5, 0x15, 0xbc, 0x99, 0x25, 0x4c, 0xce, 0x44, 0x1c,
	0xba, 0xdb, 0x1d, 0xab, 0x6b, 0xd1, 0x15, 0x61, 0x3a, 0x48, 0x33, 0x7e, 0x13, 0xcd, 0x99, 0xfb,
	0xa4, 0xe8, 0xa0, 0x61, 0x48, 0x07, 0x5a, 0xa7, 0x82, 0x4f, 0xb2, 0x24, 0x61, 0x7c, 0xb2, 0x74,
	0x77, 0xd0, 0xa1, 0x4c, 0x79, 0x7f, 0xd8, 0x78, 0xc7, 0xc7, 0xa9, 0x9f, 0x66, 0xf2, 0xa3, 0xfa,
	0xe7, 0xc1, 0x36, 0x0a, 0x49, 0x4a, 0xad, 0x7c, 0x1b, 0x6f, 0xea, 0x1a, 0x87, 0x7d, 0x49, 0x12,
	0x91, 0x68, 0x0f, 0x3d, 0x1e, 0x4a, 0x0c, 0xe9, 0xc2, 0xee, 0x85, 0x2f, 0xd3, 0x1f, 0x7c, 0x1e,
	0xc6, 0xac, 0x3f, 0x65, 0xe6, 0x02, 0x39, 0xf4, 0x21, 0xad, 0x32, 0x51, 0x94, 0xd9, 0xdd, 0xb4,
	0xb8, 0x4c, 0xa9, 0x4a, 0x29, 0x88, 0xbb, 0x63, 0x8b, 0x1d, 0xba, 0x22, 0xc8, 0x8f, 0xf0, 0x84,
	0x66, 0x9c, 0x47, 0x7c, 0xaa, 0x53, 0x75, 0x9b, 0x28, 0xea, 0xcf, 0xcb, 0x75, 0xd7, 0x96, 0xde,
	0x9a, 0xdf, 0x90, 0xa7, 0xc9, 0x92, 0xae, 0xaf, 0x6d, 0xff, 0x0c, 0x64, 0xd3, 0x89, 0xec, 0x81,
	0x7d, 0xc7, 0x96, 0x58, 0x3b, 0x9b, 0xaa, 0x4f, 0xf2, 0x1a, 0x6a, 0xef, 0xfd, 0x38, 0xd3, 0x73,
	0xb5, 0x75, 0x72, 0x50, 0x1c, 0xb6, 0xb6, 0x9a, 0x6a, 0xa7, 0xef, 0x2a, 0xdf, 0x5a, 0xde, 0xe8,
	0x41, 0x98, 0x98, 0x15, 0x7b, 0x97, 0xe2, 0xc8, 0xc4, 0xad, 0x6b, 0x74, 0x45, 0xa8, 0xaa, 0x18,
	0x77, 0x25, 0x41, 0x33, 0xc8, 0xca, 0x94, 0xf7, 0x57, 0x05, 0x9c, 0x62, 0x1e, 0x3f, 0xae, 0x66,
	0xeb, 0x43, 0x6a, 0x3e, 0x80, 0xfa, 0x28, 0x4b, 0x17, 0x59, 0x8a, 0xf1, 0x6f, 0x53, 0x83, 0xca,
	0x13, 0xc5, 0x5e, 0x9f, 0x28, 0x87, 0xe0, 0x8c, 0x53, 0x3f, 0x49, 0x51, 0x8e, 0xba, 0xdd, 0x2b,
	0x42, 0xa9, 0xe1, 0x4d, 0xc4, 0x23, 0x39, 0x43, 0x73, 0x4d, 0xab, 0x61, 0xc5, 0x14, 0xea, 0xab,
	0x3f, 0xaa, 0xbe, 0xc6, 0x9a, 0xfa, 0x5c, 0x68, 0x50, 0x96, 0x26, 0x11, 0x93, 0x78, 0x63, 0x6d,
	0x9a, 0x43, 0x35, 0x3b, 0xf4, 0x6b, 0xe4, 0x20, 0xaf, 0x81, 0xa9, 0x54, 0x91, 0xb3, 0xbe, 0x97,
	0x65, 0xca, 0xfb, 0xc7, 0x82, 0xaa, 0x02, 0x1f, 0x75, 0x09, 0x5c, 0x68, 0xfc, 0xc4, 0x12, 0xb9,
	0x7a, 0x1f, 0x73, 0xa8, 0x56, 0xf4, 0xe3, 0xe8, 0x3d, 0x0b, 0xcd, 0x0c, 0x33, 0x48, 0xed, 0x4e,
	0x45, 0xcc, 0x8c, 0xce, 0xf1, 0x9b, 0xbc, 0x82, 0xea, 0x8d, 0x3f, 0x55, 0xaa, 0x56, 0x9a, 0xfc,
	0xb4, 0x90, 0x89, 0x0a, 0xa7, 0xa7, 0x2c, 0x5a, 0x85, 0xe8, 0xd4, 0xfe, 0x06, 0x9c, 0x82, 0x2a,
	0x6b, 0xce, 0xd1, 0x9a, 0xdb, 0x2f, 0x6b, 0xce, 0x29, 0x69, 0xeb, 0xe5, 0x97, 0xf8, 0xb6, 0x90,
	0x26, 0x54, 0xe9, 0xb0, 0x3f, 0xd8, 0xdb, 0x22, 0x0d, 0xb0, 0xfb, 0x83, 0xc1, 0x9e, 0x45, 0x00,
	0xea, 0x97, 0xa3, 0xc1, 0xd9, 0x9b, 0x5f, 0xf6, 0x2a, 0xea, 0x7b, 0x30, 0xbc, 0x18, 0xde, 0x0c,
	0xf7, 0xec, 0x93, 0x7f, 0x2d, 0xb0, 0x6f, 0x45, 0x40, 0x2e, 0x61, 0xe7, 0x2d, 0x5b, 0xfb, 0x9f,
	0x78, 0xbe, 0x92, 0xf2, 0xc6, 0xcf, 0x46, 0xfb, 0xb0, 0x64, 0xdc, 0x78, 0xf7, 0xbd, 0x2d, 0x72,
	0x06, 0xcf, 0xc6, 0x8c, 0x87, 0xdf, 0xfb, 0x93, 0xbb, 0x42, 0x9a, 0xc5, 0xce, 0xa4, 0x58, 0x5c,
	0xd8, 0xda, 0x07, 0x1b, 0x4f, 0xcf, 0x50, 0xfd, 0x0d, 0x79, 0x5b, 0xe4, 0x15, 0xd4, 0x06, 0x42,
	0x65, 0xb5, 0x7a, 0x08, 0xf5, 0x8b, 0xdb, 0xde, 0x2d, 0x07, 0x91, 0xc5, 0xa9, 0xb7, 0x45, 0xbe,
	0x82, 0xd6, 0x75, 0x22, 0xee, 0x97, 0x6a, 0xc4, 0x66, 0x9c, 0xac, 0xcd, 0xdc, 0xf6, 0x23, 0xe7,
	0x7a, 0x5b, 0x41, 0x1d, 0x4f, 0xfc, 0xfa, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xac, 0x56, 0x87,
	0x67, 0xab, 0x09, 0x00, 0x00,
}
