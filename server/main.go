package main

import (
	"fmt"
	"net/http"

	"github.com/xupin/protocol-demo/network/ws"
	"github.com/xupin/protocol-demo/pb"
	"github.com/xupin/protocol-demo/utils/packet"

	"google.golang.org/protobuf/proto"
)

type Agent struct {
	Id   string
	Conn *ws.Connection
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
		Id:   conn.GetConnID(),
		Conn: conn,
	}
	for {
		// 读取消息
		msg, err := agent.Conn.Receive()
		if err != nil {
			break
		}
		switch msg.MessageType {
		case ws.BinaryMessage:
			cmd, p := packet.Decode(msg.Data)
			if cmd != "C2S_Login" {
				fmt.Println("非法协议", cmd)
				continue
			}
			req := &pb.C2S_Login{}
			proto.Unmarshal(p, req)
			fmt.Printf("登录用户: %s \n", req.Username)

			resp := &pb.S2C_UserInfo{
				Username: req.Username,
				Message:  "你好, " + req.Username,
			}
			agent.Send("S2C_UserInfo", resp)
		default:
			fmt.Printf("消息类型不支持: %d \n", msg.MessageType)
		}
	}
}

func (r *Agent) Send(cmd string, resp proto.Message) {
	p, _ := proto.Marshal(resp)
	r.Conn.Write(&ws.Message{
		MessageType: ws.BinaryMessage,
		Data:        packet.Encode(cmd, p),
	})
}
