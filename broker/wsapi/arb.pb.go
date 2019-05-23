// Code generated by protoc-gen-go. DO NOT EDIT.
// source: broker/wsapi/arb.proto

package wsapi

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ArbMarket struct {
	HePair               string                  `protobuf:"bytes,1,opt,name=he_pair,json=hePair,proto3" json:"he_pair,omitempty"`
	Spread               float64                 `protobuf:"fixed64,2,opt,name=spread,proto3" json:"spread,omitempty"`
	Low                  *ArbMarket_ActiveMarket `protobuf:"bytes,3,opt,name=low,proto3" json:"low,omitempty"`
	High                 *ArbMarket_ActiveMarket `protobuf:"bytes,4,opt,name=high,proto3" json:"high,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *ArbMarket) Reset()         { *m = ArbMarket{} }
func (m *ArbMarket) String() string { return proto.CompactTextString(m) }
func (*ArbMarket) ProtoMessage()    {}
func (*ArbMarket) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb5fe075d6d1fdf7, []int{0}
}

func (m *ArbMarket) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ArbMarket.Unmarshal(m, b)
}
func (m *ArbMarket) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ArbMarket.Marshal(b, m, deterministic)
}
func (m *ArbMarket) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ArbMarket.Merge(m, src)
}
func (m *ArbMarket) XXX_Size() int {
	return xxx_messageInfo_ArbMarket.Size(m)
}
func (m *ArbMarket) XXX_DiscardUnknown() {
	xxx_messageInfo_ArbMarket.DiscardUnknown(m)
}

var xxx_messageInfo_ArbMarket proto.InternalMessageInfo

func (m *ArbMarket) GetHePair() string {
	if m != nil {
		return m.HePair
	}
	return ""
}

func (m *ArbMarket) GetSpread() float64 {
	if m != nil {
		return m.Spread
	}
	return 0
}

func (m *ArbMarket) GetLow() *ArbMarket_ActiveMarket {
	if m != nil {
		return m.Low
	}
	return nil
}

func (m *ArbMarket) GetHigh() *ArbMarket_ActiveMarket {
	if m != nil {
		return m.High
	}
	return nil
}

type ArbMarket_ActiveMarket struct {
	Exchange             string   `protobuf:"bytes,1,opt,name=exchange,proto3" json:"exchange,omitempty"`
	HePair               string   `protobuf:"bytes,2,opt,name=he_pair,json=hePair,proto3" json:"he_pair,omitempty"`
	ExPair               string   `protobuf:"bytes,3,opt,name=ex_pair,json=exPair,proto3" json:"ex_pair,omitempty"`
	Price                string   `protobuf:"bytes,4,opt,name=price,proto3" json:"price,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ArbMarket_ActiveMarket) Reset()         { *m = ArbMarket_ActiveMarket{} }
func (m *ArbMarket_ActiveMarket) String() string { return proto.CompactTextString(m) }
func (*ArbMarket_ActiveMarket) ProtoMessage()    {}
func (*ArbMarket_ActiveMarket) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb5fe075d6d1fdf7, []int{0, 0}
}

func (m *ArbMarket_ActiveMarket) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ArbMarket_ActiveMarket.Unmarshal(m, b)
}
func (m *ArbMarket_ActiveMarket) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ArbMarket_ActiveMarket.Marshal(b, m, deterministic)
}
func (m *ArbMarket_ActiveMarket) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ArbMarket_ActiveMarket.Merge(m, src)
}
func (m *ArbMarket_ActiveMarket) XXX_Size() int {
	return xxx_messageInfo_ArbMarket_ActiveMarket.Size(m)
}
func (m *ArbMarket_ActiveMarket) XXX_DiscardUnknown() {
	xxx_messageInfo_ArbMarket_ActiveMarket.DiscardUnknown(m)
}

var xxx_messageInfo_ArbMarket_ActiveMarket proto.InternalMessageInfo

func (m *ArbMarket_ActiveMarket) GetExchange() string {
	if m != nil {
		return m.Exchange
	}
	return ""
}

func (m *ArbMarket_ActiveMarket) GetHePair() string {
	if m != nil {
		return m.HePair
	}
	return ""
}

func (m *ArbMarket_ActiveMarket) GetExPair() string {
	if m != nil {
		return m.ExPair
	}
	return ""
}

func (m *ArbMarket_ActiveMarket) GetPrice() string {
	if m != nil {
		return m.Price
	}
	return ""
}

func init() {
	proto.RegisterType((*ArbMarket)(nil), "wsapi.ArbMarket")
	proto.RegisterType((*ArbMarket_ActiveMarket)(nil), "wsapi.ArbMarket.ActiveMarket")
}

func init() { proto.RegisterFile("broker/wsapi/arb.proto", fileDescriptor_fb5fe075d6d1fdf7) }

var fileDescriptor_fb5fe075d6d1fdf7 = []byte{
	// 216 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x41, 0x4a, 0xc4, 0x30,
	0x18, 0x85, 0x49, 0x3b, 0xd3, 0xb1, 0xbf, 0xae, 0x82, 0xcc, 0x84, 0x01, 0xa1, 0xb8, 0xea, 0x2a,
	0x45, 0x3d, 0xc1, 0x1c, 0x40, 0x90, 0x5c, 0x40, 0x92, 0xfa, 0xd3, 0x84, 0x8a, 0x09, 0x7f, 0x8b,
	0xed, 0x41, 0x3c, 0xb0, 0x98, 0x94, 0x52, 0x77, 0xb3, 0xfc, 0xfe, 0xf7, 0xf2, 0x78, 0x2f, 0x70,
	0x34, 0xe4, 0x7b, 0xa4, 0x66, 0x1a, 0x74, 0x70, 0x8d, 0x26, 0x23, 0x03, 0xf9, 0xd1, 0xf3, 0x7d,
	0x3c, 0x3c, 0xfe, 0x64, 0x50, 0x5e, 0xc8, 0xbc, 0x6a, 0xea, 0x71, 0xe4, 0x27, 0x38, 0x58, 0x7c,
	0x0f, 0xda, 0x91, 0x60, 0x15, 0xab, 0x4b, 0x55, 0x58, 0x7c, 0xd3, 0x8e, 0xf8, 0x11, 0x8a, 0x21,
	0x10, 0xea, 0x0f, 0x91, 0x55, 0xac, 0x66, 0x6a, 0x21, 0xde, 0x40, 0xfe, 0xe9, 0x27, 0x91, 0x57,
	0xac, 0xbe, 0x7d, 0x7e, 0x90, 0x31, 0x53, 0xae, 0x79, 0xf2, 0xd2, 0x8e, 0xee, 0x1b, 0x13, 0xa8,
	0x3f, 0x27, 0x7f, 0x82, 0x9d, 0x75, 0x9d, 0x15, 0xbb, 0x6b, 0x5e, 0x44, 0xeb, 0x99, 0xe0, 0x6e,
	0x7b, 0xe5, 0x67, 0xb8, 0xc1, 0xb9, 0xb5, 0xfa, 0xab, 0xc3, 0xa5, 0xe5, 0xca, 0xdb, 0x01, 0xd9,
	0xbf, 0x01, 0x27, 0x38, 0xe0, 0x9c, 0x84, 0x3c, 0x09, 0x38, 0x47, 0xe1, 0x1e, 0xf6, 0x81, 0x5c,
	0x8b, 0xb1, 0x51, 0xa9, 0x12, 0x98, 0x22, 0x7e, 0xd2, 0xcb, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x20, 0x3d, 0x27, 0x08, 0x3e, 0x01, 0x00, 0x00,
}
