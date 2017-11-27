// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message/api.proto

/*
Package message is a generated protocol buffer package.

It is generated from these files:
	message/api.proto
	message/serfQueryParams.proto
	message/job.proto

It has these top-level messages:
	ApiJobResponse
	ApiExecutionResponse
	ApiJobStatusResponse
	ApiStringResponse
	ApiNodeResponse
	Pageing
	SearchCondition
	ApiJobQueryString
	ApiJobStatusQueryString
	ApiExecutionQueryString
	ApiSearchQueryString
	ApiNodeQueryString
	JobQueryParams
	GetRPCConfigResp
	JobCountResp
	QueryResult
	QueryJobRunParams
	Search
	Params
	Result
	Job
	JobStatus
	Execution
	Node
*/
package message

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ApiJobResponse struct {
	Succeed    bool   `protobuf:"varint,1,opt,name=Succeed" json:"Succeed,omitempty"`
	Message    string `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
	MaxPageNum int32  `protobuf:"varint,3,opt,name=MaxPageNum" json:"MaxPageNum,omitempty"`
	Data       []*Job `protobuf:"bytes,4,rep,name=Data" json:"Data,omitempty"`
}

func (m *ApiJobResponse) Reset()                    { *m = ApiJobResponse{} }
func (m *ApiJobResponse) String() string            { return proto.CompactTextString(m) }
func (*ApiJobResponse) ProtoMessage()               {}
func (*ApiJobResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ApiJobResponse) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *ApiJobResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *ApiJobResponse) GetMaxPageNum() int32 {
	if m != nil {
		return m.MaxPageNum
	}
	return 0
}

func (m *ApiJobResponse) GetData() []*Job {
	if m != nil {
		return m.Data
	}
	return nil
}

type ApiExecutionResponse struct {
	Succeed    bool         `protobuf:"varint,1,opt,name=Succeed" json:"Succeed,omitempty"`
	Message    string       `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
	MaxPageNum int32        `protobuf:"varint,3,opt,name=MaxPageNum" json:"MaxPageNum,omitempty"`
	Data       []*Execution `protobuf:"bytes,4,rep,name=Data" json:"Data,omitempty"`
}

func (m *ApiExecutionResponse) Reset()                    { *m = ApiExecutionResponse{} }
func (m *ApiExecutionResponse) String() string            { return proto.CompactTextString(m) }
func (*ApiExecutionResponse) ProtoMessage()               {}
func (*ApiExecutionResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ApiExecutionResponse) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *ApiExecutionResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *ApiExecutionResponse) GetMaxPageNum() int32 {
	if m != nil {
		return m.MaxPageNum
	}
	return 0
}

func (m *ApiExecutionResponse) GetData() []*Execution {
	if m != nil {
		return m.Data
	}
	return nil
}

type ApiJobStatusResponse struct {
	Succeed    bool         `protobuf:"varint,1,opt,name=Succeed" json:"Succeed,omitempty"`
	Message    string       `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
	MaxPageNum int32        `protobuf:"varint,3,opt,name=MaxPageNum" json:"MaxPageNum,omitempty"`
	Data       []*JobStatus `protobuf:"bytes,4,rep,name=Data" json:"Data,omitempty"`
}

func (m *ApiJobStatusResponse) Reset()                    { *m = ApiJobStatusResponse{} }
func (m *ApiJobStatusResponse) String() string            { return proto.CompactTextString(m) }
func (*ApiJobStatusResponse) ProtoMessage()               {}
func (*ApiJobStatusResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ApiJobStatusResponse) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *ApiJobStatusResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *ApiJobStatusResponse) GetMaxPageNum() int32 {
	if m != nil {
		return m.MaxPageNum
	}
	return 0
}

func (m *ApiJobStatusResponse) GetData() []*JobStatus {
	if m != nil {
		return m.Data
	}
	return nil
}

type ApiStringResponse struct {
	Succeed    bool     `protobuf:"varint,1,opt,name=Succeed" json:"Succeed,omitempty"`
	Message    string   `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
	MaxPageNum int32    `protobuf:"varint,3,opt,name=MaxPageNum" json:"MaxPageNum,omitempty"`
	Data       []string `protobuf:"bytes,4,rep,name=Data" json:"Data,omitempty"`
}

func (m *ApiStringResponse) Reset()                    { *m = ApiStringResponse{} }
func (m *ApiStringResponse) String() string            { return proto.CompactTextString(m) }
func (*ApiStringResponse) ProtoMessage()               {}
func (*ApiStringResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ApiStringResponse) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *ApiStringResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *ApiStringResponse) GetMaxPageNum() int32 {
	if m != nil {
		return m.MaxPageNum
	}
	return 0
}

func (m *ApiStringResponse) GetData() []string {
	if m != nil {
		return m.Data
	}
	return nil
}

type ApiNodeResponse struct {
	Succeed    bool    `protobuf:"varint,1,opt,name=Succeed" json:"Succeed,omitempty"`
	Message    string  `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
	MaxPageNum int32   `protobuf:"varint,3,opt,name=MaxPageNum" json:"MaxPageNum,omitempty"`
	Data       []*Node `protobuf:"bytes,4,rep,name=Data" json:"Data,omitempty"`
}

func (m *ApiNodeResponse) Reset()                    { *m = ApiNodeResponse{} }
func (m *ApiNodeResponse) String() string            { return proto.CompactTextString(m) }
func (*ApiNodeResponse) ProtoMessage()               {}
func (*ApiNodeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ApiNodeResponse) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

func (m *ApiNodeResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *ApiNodeResponse) GetMaxPageNum() int32 {
	if m != nil {
		return m.MaxPageNum
	}
	return 0
}

func (m *ApiNodeResponse) GetData() []*Node {
	if m != nil {
		return m.Data
	}
	return nil
}

type Pageing struct {
	// @inject_tag: form:"pagenum"
	PageNum int32 `protobuf:"varint,1,opt,name=PageNum" json:"PageNum,omitempty" form:"pagenum"`
	// @inject_tag: form:"pagesize"
	PageSize int32 `protobuf:"varint,2,opt,name=PageSize" json:"PageSize,omitempty" form:"pagesize"`
	// @inject_tag: form:"maxpage"
	OutMaxPage bool `protobuf:"varint,3,opt,name=OutMaxPage" json:"OutMaxPage,omitempty" form:"maxpage"`
}

func (m *Pageing) Reset()                    { *m = Pageing{} }
func (m *Pageing) String() string            { return proto.CompactTextString(m) }
func (*Pageing) ProtoMessage()               {}
func (*Pageing) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Pageing) GetPageNum() int32 {
	if m != nil {
		return m.PageNum
	}
	return 0
}

func (m *Pageing) GetPageSize() int32 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

func (m *Pageing) GetOutMaxPage() bool {
	if m != nil {
		return m.OutMaxPage
	}
	return false
}

type SearchCondition struct {
	// @inject_tag: form:"conditions"
	Conditions []string `protobuf:"bytes,1,rep,name=Conditions" json:"Conditions,omitempty" form:"conditions"`
	// @inject_tag: form:"links"
	Links []string `protobuf:"bytes,2,rep,name=Links" json:"Links,omitempty" form:"links"`
}

func (m *SearchCondition) Reset()                    { *m = SearchCondition{} }
func (m *SearchCondition) String() string            { return proto.CompactTextString(m) }
func (*SearchCondition) ProtoMessage()               {}
func (*SearchCondition) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *SearchCondition) GetConditions() []string {
	if m != nil {
		return m.Conditions
	}
	return nil
}

func (m *SearchCondition) GetLinks() []string {
	if m != nil {
		return m.Links
	}
	return nil
}

type ApiJobQueryString struct {
	// @inject_tag: form:"job"
	Job *Job `protobuf:"bytes,1,opt,name=Job" json:"Job,omitempty" form:"job"`
	// @inject_tag: form:"pageing"
	Pageing *Pageing `protobuf:"bytes,2,opt,name=Pageing" json:"Pageing,omitempty" form:"pageing"`
}

func (m *ApiJobQueryString) Reset()                    { *m = ApiJobQueryString{} }
func (m *ApiJobQueryString) String() string            { return proto.CompactTextString(m) }
func (*ApiJobQueryString) ProtoMessage()               {}
func (*ApiJobQueryString) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *ApiJobQueryString) GetJob() *Job {
	if m != nil {
		return m.Job
	}
	return nil
}

func (m *ApiJobQueryString) GetPageing() *Pageing {
	if m != nil {
		return m.Pageing
	}
	return nil
}

type ApiJobStatusQueryString struct {
	// @inject_tag: form:"status"
	Status *JobStatus `protobuf:"bytes,1,opt,name=Status" json:"Status,omitempty" form:"status"`
	// @inject_tag: form:"pageing"
	Pageing *Pageing `protobuf:"bytes,2,opt,name=Pageing" json:"Pageing,omitempty" form:"pageing"`
}

func (m *ApiJobStatusQueryString) Reset()                    { *m = ApiJobStatusQueryString{} }
func (m *ApiJobStatusQueryString) String() string            { return proto.CompactTextString(m) }
func (*ApiJobStatusQueryString) ProtoMessage()               {}
func (*ApiJobStatusQueryString) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *ApiJobStatusQueryString) GetStatus() *JobStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *ApiJobStatusQueryString) GetPageing() *Pageing {
	if m != nil {
		return m.Pageing
	}
	return nil
}

type ApiExecutionQueryString struct {
	// @inject_tag: form:"execution"
	Execution *Execution `protobuf:"bytes,1,opt,name=Execution" json:"Execution,omitempty" form:"execution"`
	// @inject_tag: form:"pageing"
	Pageing *Pageing `protobuf:"bytes,2,opt,name=Pageing" json:"Pageing,omitempty" form:"pageing"`
}

func (m *ApiExecutionQueryString) Reset()                    { *m = ApiExecutionQueryString{} }
func (m *ApiExecutionQueryString) String() string            { return proto.CompactTextString(m) }
func (*ApiExecutionQueryString) ProtoMessage()               {}
func (*ApiExecutionQueryString) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *ApiExecutionQueryString) GetExecution() *Execution {
	if m != nil {
		return m.Execution
	}
	return nil
}

func (m *ApiExecutionQueryString) GetPageing() *Pageing {
	if m != nil {
		return m.Pageing
	}
	return nil
}

type ApiSearchQueryString struct {
	// @inject_tag: form:"q"
	SearchCondition *SearchCondition `protobuf:"bytes,1,opt,name=SearchCondition" json:"SearchCondition,omitempty" form:"q"`
	// @inject_tag: form:"pageing"
	Pageing *Pageing `protobuf:"bytes,2,opt,name=Pageing" json:"Pageing,omitempty" form:"pageing"`
}

func (m *ApiSearchQueryString) Reset()                    { *m = ApiSearchQueryString{} }
func (m *ApiSearchQueryString) String() string            { return proto.CompactTextString(m) }
func (*ApiSearchQueryString) ProtoMessage()               {}
func (*ApiSearchQueryString) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *ApiSearchQueryString) GetSearchCondition() *SearchCondition {
	if m != nil {
		return m.SearchCondition
	}
	return nil
}

func (m *ApiSearchQueryString) GetPageing() *Pageing {
	if m != nil {
		return m.Pageing
	}
	return nil
}

type ApiNodeQueryString struct {
	// @inject_tag: form:"node"
	Node *Node `protobuf:"bytes,1,opt,name=Node" json:"Node,omitempty" form:"node"`
	// @inject_tag: form:"pageing"
	Pageing *Pageing `protobuf:"bytes,2,opt,name=Pageing" json:"Pageing,omitempty" form:"pageing"`
}

func (m *ApiNodeQueryString) Reset()                    { *m = ApiNodeQueryString{} }
func (m *ApiNodeQueryString) String() string            { return proto.CompactTextString(m) }
func (*ApiNodeQueryString) ProtoMessage()               {}
func (*ApiNodeQueryString) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *ApiNodeQueryString) GetNode() *Node {
	if m != nil {
		return m.Node
	}
	return nil
}

func (m *ApiNodeQueryString) GetPageing() *Pageing {
	if m != nil {
		return m.Pageing
	}
	return nil
}

func init() {
	proto.RegisterType((*ApiJobResponse)(nil), "message.ApiJobResponse")
	proto.RegisterType((*ApiExecutionResponse)(nil), "message.ApiExecutionResponse")
	proto.RegisterType((*ApiJobStatusResponse)(nil), "message.ApiJobStatusResponse")
	proto.RegisterType((*ApiStringResponse)(nil), "message.ApiStringResponse")
	proto.RegisterType((*ApiNodeResponse)(nil), "message.ApiNodeResponse")
	proto.RegisterType((*Pageing)(nil), "message.Pageing")
	proto.RegisterType((*SearchCondition)(nil), "message.SearchCondition")
	proto.RegisterType((*ApiJobQueryString)(nil), "message.ApiJobQueryString")
	proto.RegisterType((*ApiJobStatusQueryString)(nil), "message.ApiJobStatusQueryString")
	proto.RegisterType((*ApiExecutionQueryString)(nil), "message.ApiExecutionQueryString")
	proto.RegisterType((*ApiSearchQueryString)(nil), "message.ApiSearchQueryString")
	proto.RegisterType((*ApiNodeQueryString)(nil), "message.ApiNodeQueryString")
}

func init() { proto.RegisterFile("message/api.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 455 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x94, 0x4d, 0x6f, 0xd3, 0x30,
	0x18, 0x80, 0xe5, 0xb5, 0x5d, 0xdb, 0x77, 0xc0, 0x98, 0x35, 0x09, 0x6b, 0x87, 0x29, 0xcb, 0x01,
	0x55, 0x3b, 0x14, 0x34, 0x7e, 0x41, 0xf8, 0x10, 0x52, 0xc4, 0x06, 0x38, 0x3f, 0x60, 0x72, 0x52,
	0x2b, 0x18, 0xb4, 0x38, 0xe4, 0x43, 0x0c, 0xb8, 0x57, 0x9c, 0xf9, 0xc5, 0xc8, 0x1f, 0x71, 0xdc,
	0x96, 0x4b, 0x0e, 0xd5, 0x2e, 0xd3, 0xde, 0x8f, 0xfa, 0x79, 0xea, 0xd7, 0x6f, 0xe1, 0xe4, 0x8e,
	0xd7, 0x35, 0xcb, 0xf9, 0x0b, 0x56, 0x8a, 0x65, 0x59, 0xc9, 0x46, 0xe2, 0xa9, 0x4d, 0x9d, 0xb9,
	0xda, 0x57, 0x99, 0x9a, 0x5a, 0xb8, 0x46, 0xf0, 0x24, 0x2a, 0x45, 0x2c, 0x53, 0xca, 0xeb, 0x52,
	0x16, 0x35, 0xc7, 0x04, 0xa6, 0x49, 0x9b, 0x65, 0x9c, 0xaf, 0x08, 0x0a, 0xd0, 0x62, 0x46, 0xbb,
	0x50, 0x55, 0xae, 0xcd, 0x09, 0xe4, 0x20, 0x40, 0x8b, 0x39, 0xed, 0x42, 0x7c, 0x0e, 0x70, 0xcd,
	0xee, 0x3f, 0xb1, 0x9c, 0xdf, 0xb4, 0x77, 0x64, 0x14, 0xa0, 0xc5, 0x84, 0x7a, 0x19, 0x1c, 0xc0,
	0xf8, 0x2d, 0x6b, 0x18, 0x19, 0x07, 0xa3, 0xc5, 0xd1, 0xd5, 0xa3, 0xa5, 0x15, 0x59, 0x2a, 0xae,
	0xae, 0x84, 0x7f, 0x11, 0x9c, 0x46, 0xa5, 0x78, 0x77, 0xcf, 0xb3, 0xb6, 0x11, 0xb2, 0xd8, 0xab,
	0xce, 0xf3, 0x0d, 0x1d, 0xec, 0x74, 0x7a, 0xfa, 0x86, 0x54, 0x2c, 0xd3, 0xa4, 0x61, 0x4d, 0x5b,
	0x3f, 0x88, 0x54, 0x4f, 0x37, 0x52, 0xbf, 0xe1, 0x24, 0x2a, 0x45, 0xd2, 0x54, 0xa2, 0xc8, 0xf7,
	0x2a, 0x84, 0x3d, 0xa1, 0xb9, 0x85, 0xff, 0x41, 0x70, 0x1c, 0x95, 0xe2, 0x46, 0xae, 0xf8, 0x5e,
	0xd9, 0x17, 0x1b, 0x97, 0xf1, 0xd8, 0x5d, 0x86, 0x06, 0x1b, 0x95, 0x5b, 0x98, 0xaa, 0x6e, 0x51,
	0xe4, 0x8a, 0xd3, 0x1d, 0x85, 0xf4, 0x51, 0x5d, 0x88, 0xcf, 0x60, 0xa6, 0xfe, 0x4d, 0xc4, 0x2f,
	0xa3, 0x30, 0xa1, 0x2e, 0x56, 0x0e, 0x1f, 0xdb, 0xc6, 0x42, 0xb5, 0xc3, 0x8c, 0x7a, 0x99, 0xf0,
	0x3d, 0x1c, 0x27, 0x9c, 0x55, 0xd9, 0x97, 0x37, 0xb2, 0x58, 0x09, 0xf5, 0x2c, 0xd4, 0x47, 0x5c,
	0x50, 0x13, 0xa4, 0x2f, 0xc6, 0xcb, 0xe0, 0x53, 0x98, 0x7c, 0x10, 0xc5, 0xb7, 0x9a, 0x1c, 0xe8,
	0x92, 0x09, 0xc2, 0x5b, 0x3d, 0xb1, 0x58, 0xa6, 0x9f, 0x5b, 0x5e, 0xfd, 0x34, 0x93, 0xc3, 0xe7,
	0x30, 0x8a, 0x65, 0xaa, 0x7d, 0xb7, 0x37, 0x42, 0x15, 0xf0, 0xa5, 0xfb, 0x7a, 0x5a, 0xfc, 0xe8,
	0xea, 0xa9, 0xeb, 0xb1, 0x79, 0xda, 0x35, 0x84, 0xdf, 0xe1, 0x99, 0xff, 0x4c, 0x7d, 0xcc, 0x25,
	0x1c, 0x9a, 0xa4, 0x25, 0xfd, 0xef, 0x5d, 0xd9, 0x8e, 0x41, 0xc8, 0x1f, 0x1a, 0xe9, 0x16, 0xc6,
	0x47, 0xbe, 0x84, 0xb9, 0xcb, 0xef, 0x50, 0xfb, 0x15, 0xeb, 0x9b, 0x06, 0x81, 0xd7, 0x66, 0x27,
	0xcd, 0x64, 0x7c, 0xec, 0xeb, 0x9d, 0x71, 0x59, 0x38, 0x71, 0x87, 0x6d, 0xd5, 0xe9, 0xce, 0x7c,
	0x87, 0x88, 0x64, 0x80, 0xed, 0x26, 0xf8, 0x16, 0x17, 0x30, 0x56, 0x29, 0x8b, 0xde, 0x7e, 0xb8,
	0xea, 0xef, 0x10, 0x48, 0x7a, 0xa8, 0x7f, 0xa6, 0x5f, 0xfd, 0x0b, 0x00, 0x00, 0xff, 0xff, 0xd2,
	0x36, 0x26, 0x3a, 0xd7, 0x05, 0x00, 0x00,
}
