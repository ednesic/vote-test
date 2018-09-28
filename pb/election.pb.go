// Code generated by protoc-gen-go. DO NOT EDIT.
// source: election.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	election.proto
	vote.proto

It has these top-level messages:
	Election
	Vote
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Election struct {
	Id      int32                      `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Inicio  *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=inicio" json:"inicio,omitempty"`
	Termino *google_protobuf.Timestamp `protobuf:"bytes,3,opt,name=termino" json:"termino,omitempty"`
}

func (m *Election) Reset()                    { *m = Election{} }
func (m *Election) String() string            { return proto.CompactTextString(m) }
func (*Election) ProtoMessage()               {}
func (*Election) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Election) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Election) GetInicio() *google_protobuf.Timestamp {
	if m != nil {
		return m.Inicio
	}
	return nil
}

func (m *Election) GetTermino() *google_protobuf.Timestamp {
	if m != nil {
		return m.Termino
	}
	return nil
}

func init() {
	proto.RegisterType((*Election)(nil), "Election")
}

func init() { proto.RegisterFile("election.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 149 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4b, 0xcd, 0x49, 0x4d,
	0x2e, 0xc9, 0xcc, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x97, 0x92, 0x4f, 0xcf, 0xcf, 0x4f,
	0xcf, 0x49, 0xd5, 0x07, 0xf3, 0x92, 0x4a, 0xd3, 0xf4, 0x4b, 0x32, 0x73, 0x53, 0x8b, 0x4b, 0x12,
	0x73, 0x0b, 0x20, 0x0a, 0x94, 0x5a, 0x18, 0xb9, 0x38, 0x5c, 0xa1, 0x7a, 0x84, 0xf8, 0xb8, 0x98,
	0x32, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35, 0x58, 0x83, 0x98, 0x32, 0x53, 0x84, 0x8c, 0xb8, 0xd8,
	0x32, 0xf3, 0x32, 0x93, 0x33, 0xf3, 0x25, 0x98, 0x14, 0x18, 0x35, 0xb8, 0x8d, 0xa4, 0xf4, 0x20,
	0xc6, 0xe9, 0xc1, 0x8c, 0xd3, 0x0b, 0x81, 0x19, 0x17, 0x04, 0x55, 0x29, 0x64, 0xc2, 0xc5, 0x5e,
	0x92, 0x5a, 0x94, 0x9b, 0x99, 0x97, 0x2f, 0xc1, 0x4c, 0x50, 0x13, 0x4c, 0xa9, 0x13, 0x4b, 0x14,
	0x53, 0x41, 0x52, 0x12, 0x1b, 0x58, 0x89, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x20, 0x51, 0xf5,
	0xf6, 0xc6, 0x00, 0x00, 0x00,
}
