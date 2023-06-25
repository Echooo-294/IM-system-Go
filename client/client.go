package main

import (
	"flag"
	"fmt"
	"im_system/public"
	"net"
	"strconv"
)

// 所连接服务器的IP
var gServeIp string

// 所连接服务器的gServePort
var gServePort int

// 包的初始化
func init() {
	// 命令行参数定义
	flag.StringVar(&gServeIp, "ip", "127.0.0.1", "设置服务器的IP地址(默认为127.0.0.1)")
	flag.IntVar(&gServePort, "port", 8888, "设置服务器的端口(默认为8888)")
}

type Client struct {
	ServeIp   string
	ServePort int
	Name      string
	conn      net.Conn

	// 当前客户端所在模式
	mode int
}

// 创建一个client
func NewClient(serveIp string, servePort int) *Client {

	client := &Client{
		ServeIp:   serveIp,
		ServePort: servePort,
		mode:      -1,
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

// 显示菜单获取用户输入
func (cli *Client) menu() bool {
	var mode int
	fmt.Println("目前功能有: ")
	modeList := [...]string{"退出", "公聊模式", "私聊模式", "更新用户名", "清屏"}

	for i, modeName := range modeList {
		fmt.Println(strconv.Itoa(i), ".", modeName)
	}

	// 读取用户输入
	fmt.Scan(&mode)

	if mode >= 0 && mode < len(modeList) {
		cli.mode = mode
		return true
	} else {
		fmt.Println("...请输入合法范围内的数字...")
		return false
	}
}

// 启动客户端业务
func (cli *Client) Run() {
	// 显示菜单等待用户输入
	for cli.mode != 0 {
		for !cli.menu() {
		}
		switch cli.mode {
		case 1:
			fmt.Println("公聊模式")
		case 2:
			fmt.Println("私聊模式")
		case 3:
			fmt.Println("更新用户名")
		default:
			public.CallClear()
		}
	}
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
	client.Run()
}
