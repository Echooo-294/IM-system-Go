package main

import (
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
}

// 创建一个client
func NewClient(serverIp string, serverPort int) *Client {

	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("net.Dial err :", err)
		return nil
	}
	client.conn = conn

	return client
}

func main() {
	// 现采用默认IP及端口,可改为用户输入
	client := NewClient("127.0.0.1", 8888)
	if client == nil {
		fmt.Println("Client creation failed to connect to the server...")
		return
	}
	fmt.Println("Client is created and successfully connects to the server!!!")

	// 启动客户端业务
	select {}
}
