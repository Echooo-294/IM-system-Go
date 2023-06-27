package main

import (
	"flag"
	"fmt"
	"im_system/public"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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

// 客户端结构体
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

// 服务器响应消息监听处理
func (c *Client) DealResponse() {
	// 持续阻塞,服务器有消息则打印输出
	io.Copy(os.Stdout, c.conn)
}

// 显示菜单获取用户输入
func (c *Client) menu() bool {
	var mode int
	fmt.Println("目前功能有: ")
	modeList := [...]string{"退出", "公聊模式", "私聊模式", "更新用户名", "清屏"}

	for i, modeName := range modeList {
		fmt.Println(strconv.Itoa(i), ".", modeName)
	}

	// 读取用户输入
	// 使用scanln会导致直接输入回车时退出程序
	fmt.Scan(&mode)

	if mode >= 0 && mode < len(modeList) {
		c.mode = mode
		return true
	} else {
		fmt.Println("...请输入合法范围内的数字...")
		return false
	}
}

// 读取用户输入
func (c *Client) readClient(tips string, limitDn int, limitUp int) (string, bool) {
	// 打印提示信息
	fmt.Println(tips)

	// 读取用户输入
	var msg string
	fmt.Scan(&msg)
	n := len(msg)

	// 输入内容为空
	if n == 0 {
		fmt.Println("输入内容不得为空.")
		return "", false
	}

	// 内容长度限制
	if n >= limitUp || n <= limitDn {
		fmt.Println("输入内容长度不符合要求.")
		return "", false
	}

	return msg, true
}

// 用户命名限制
func (c *Client) nameLimit(newName string) bool {
	// 不得与当前用户名相同
	if newName == c.Name {
		fmt.Println("用户名不得与当前用户名相同,请重新尝试.")
		return false
	}

	// 不能有空格
	return !strings.Contains(newName, " ")
}

// 更新用户名
func (c *Client) UpdateName() {
	// 读取用户输入
	newName, ok := c.readClient("请输入新用户名(不能有空格,3-20字符): ", 3, 20)
	if !ok {
		return
	}

	// 用户命名限制
	allow := c.nameLimit(newName)
	if !allow {
		fmt.Println("用户名不符合规范,请重新尝试.")
		return
	}

	order := "rename-" + newName + "\n"
	_, err := c.conn.Write([]byte(order))
	if err != nil {
		fmt.Println("Conn Write has err(UpdateName): ", err)
		return
	}

	// 应在确认修改好服务器用户名后再修改当前用户名
	c.Name = newName
}

// 私聊功能
func (c *Client) PrivateChat() {
	// 读取用户输入
	tName, ok := c.readClient("请输入私聊对象的用户名: ", 3, 20)
	if !ok {
		return
	}

	// 判断是否是给自己发送
	if tName == c.Name {
		fmt.Println("不得与自己聊天,请重新尝试.")
		return
	}

	tMsg, allow := c.readClient("请输入要发送的内容: ", 1, public.UsrMsgMaxLen)
	if !allow {
		return
	}

	order := "to-" + tName + "\n" + tMsg + "\n"
	_, err := c.conn.Write([]byte(order))
	if err != nil {
		fmt.Println("Conn Write has err(PrivateChat): ", err)
		return
	}

	// 如何确认消息已成功发送
}

// 退出客户端
func (c *Client) exitClient() {

}

// 启动客户端业务
func (c *Client) Run() {
	// 显示菜单等待用户输入
	for c.mode != 0 {
		for !c.menu() {
		}
		switch c.mode {
		case 0:
			fmt.Println("-退出-")
			c.exitClient()
		case 1:
			fmt.Println("-公聊模式-")
		case 2:
			fmt.Println("-私聊模式-")
		case 3:
			fmt.Println("-更新用户名-")
			c.UpdateName()
		case 4:
			public.CallClear()
		default:
			c.mode = -1
		}
		// 设置延时,避免屏幕刷新比菜单显示慢
		time.Sleep(time.Duration(50 * time.Millisecond))
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

	// 启动服务器响应消息监听处理
	go client.DealResponse()

	// 启动客户端业务
	client.Run()
}
