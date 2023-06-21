package main

import (
	"flag"
	"fmt"
	"net"
)

// 所连接服务器的IP
var gServeIp string

// 所连接服务器的gServePort
var gServePort int

type Client struct {
	ServeIp   string
	ServePort int
	Name      string
	conn      net.Conn
}

// 创建一个client
func NewClient(serveIp string, servePort int) *Client {

	client := &Client{
		ServeIp:   serveIp,
		ServePort: servePort,
	}

	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServeIp, client.ServePort))
	if err != nil {
		fmt.Println("net.Dial err :", err)
		return nil
	}
	client.conn = conn

	return client
}

// 包的初始化
func init() {
	// 命令行参数定义
	flag.StringVar(&gServeIp, "ip", "127.0.0.1", "设置服务器的IP地址(默认为127.0.0.1)")
	flag.IntVar(&gServePort, "port", 8888, "设置服务器的端口(默认为8888)")
}

func main() {
	// 运行程序后，启用命令行解析
	flag.Parse()
	client := NewClient(gServeIp, gServePort)
	if client == nil {
		fmt.Println("Client creation failed to connect to the server...")
		return
	}
	fmt.Println("Client is created and successfully connects to the server!!!")

	// 启动客户端业务
	select {}
}
