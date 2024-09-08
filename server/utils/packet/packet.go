package packet

import (
	"bytes"
	"encoding/binary"
)

func Encode(cmd string, p []byte) []byte {
	buffer := &bytes.Buffer{}
	// 2字节，协议名称长度
	buffer.Write(shortBytes(uint16(len(cmd))))
	// 协议名称
	buffer.WriteString(cmd)
	// 协议内容
	buffer.Write(p)
	return buffer.Bytes()
}

func Decode(p []byte) (string, []byte) {
	buffer := bytes.NewBuffer(p)
	// 2字节，协议名称长度
	headLen := make([]byte, 2)
	buffer.Read(headLen)
	// 协议名称
	cmdLen := readShort(headLen)
	cmd := make([]byte, cmdLen)
	buffer.Read(cmd)
	// 协议内容
	var bytes []byte
	if l := buffer.Len(); l > 0 {
		bytes = make([]byte, l)
		buffer.Read(bytes)
	}
	return string(cmd), bytes
}

func readShort(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func shortBytes(i uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, i)
	return bytes
}
