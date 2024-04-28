package network

import (
	"bytes"
	"encoding/binary"
)

type Packet struct {
	Protocol string
	Bytes    []byte
}

func (r *Packet) Encode() []byte {
	buffer := &bytes.Buffer{}
	// 2字节，协议名称长度
	buffer.Write(r.shortBytes(uint16(len(r.Protocol))))
	// 协议名称
	buffer.WriteString(r.Protocol)
	// 协议内容
	buffer.Write(r.Bytes)
	return buffer.Bytes()
}

func (r *Packet) Decode() []byte {
	buffer := bytes.NewBuffer(r.Bytes)
	byteLen := buffer.Len()
	// 2字节，协议名称长度
	headLen := make([]byte, 2)
	buffer.Read(headLen)
	// 协议名称
	protocolLen := r.readShort(headLen)
	protocol := make([]byte, protocolLen)
	buffer.Read(protocol)
	r.Protocol = string(protocol)
	// 协议内容
	bytes := make([]byte, byteLen-2-int(protocolLen))
	buffer.Read(bytes)
	r.Bytes = bytes
	return bytes
}

func (r *Packet) readShort(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func (r *Packet) shortBytes(i uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, i)
	return bytes
}
