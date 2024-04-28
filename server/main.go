package main

import (
	"fmt"
	"game-protocol/network"
	"game-protocol/network/ws"
	"game-protocol/protocol"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type Agent struct {
	Id     string
	Socket *ws.Connection
	Packet *network.Packet
	Player int32
}

func main() {
	// 开启ws服务
	http.HandleFunc("/", wsHandler)
	_ = http.ListenAndServe("0.0.0.0:9501", nil)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// 新建连接实例
	conn := ws.NewConnection()
	// 开启连接
	if err := conn.Open(w, r); err != nil {
		return
	}
	// 关闭连接
	defer conn.Close()
	agent := &Agent{
		Id:     conn.GetConnID(),
		Socket: conn,
	}
	for {
		// 读取消息
		msg, err := agent.Socket.Receive()
		if err != nil {
			break
		}
		switch msg.MessageType {
		case ws.TextMessage:
			conn.Write(&ws.Message{
				MessageType: ws.TextMessage,
				Data:        msg.Data,
			})
			conn.KeepHeartbeat()
		case ws.BinaryMessage:
			// 解包
			packet := &network.Packet{
				Bytes: msg.Data,
			}
			packet.Decode()
			// 触发协议函数
			agent.Packet = packet
			agent.Receive()
		default:
			fmt.Printf("消息类型不支持: %d \n", msg.MessageType)
		}
	}
}

func (r *Agent) Receive() {
	fmt.Printf("触发协议: %s \n", r.Packet.Protocol)
	// 接收
	pb := &protocol.Login{}
	proto.Unmarshal(r.Packet.Bytes, pb)
	fmt.Printf("登录用户: %s \n", pb.Username)
	// 发送
	pb1 := &protocol.UserInfo{
		Username: pb.Username,
		Message:  "你好, " + pb.Username,
	}
	bytes, _ := proto.Marshal(pb1)
	r.Send("user_info", bytes)
}

func (r *Agent) Send(protocol string, bytes []byte) {
	packet := network.Packet{
		Protocol: protocol,
		Bytes:    bytes,
	}
	r.Socket.Write(&ws.Message{
		MessageType: ws.BinaryMessage,
		Data:        packet.Encode(),
	})
}
