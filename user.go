package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type User struct {
	Name    string
	Address string
	Chan_u  chan string
	conn_u  net.Conn
	server  *Server
}

// 创建一个user
func NewUser(conn net.Conn, server *Server) *User {
	// 获取当前地址
	useraddr := conn.RemoteAddr().String()

	usr := &User{
		Name:    useraddr,
		Address: useraddr,
		Chan_u:  make(chan string),
		conn_u:  conn,
		server:  server,
	}

	// 启动监听当前user channel的goroutine
	go usr.ListenUsrMsg()

	return usr
}

// 监听当前user channel，有消息就发给Client
func (usr *User) ListenUsrMsg() {
	for msg := range usr.Chan_u {
		usr.conn_u.Write([]byte(msg + "\n"))
	}
}

// 给当前客户端发送消息
func (usr *User) SendMsgToClient(msg string) {
	usr.conn_u.Write([]byte(msg))
}

// 用户上线
func (usr *User) Online() {
	usr.server.RegisterUsr(usr)
	usr.server.BroadcastUsrMsg(usr, "is Online!")
}

// 用户析构关闭资源,统一在server的Handler中defer析构
func (usr *User) CloseResources() {
	usr.server.DeleteUsr(usr)
	close(usr.Chan_u)
	usr.conn_u.Close()
}

// 用户下线
func (usr *User) Offline() {
	usr.server.BroadcastUsrMsg(usr, "is Offline~")
}

// 用户被强制踢出
func (usr *User) ForceOffline() {
	usr.SendMsgToClient("You are ForceOffline.\n")
}

// num指令，查询当前在线用户人数
func (usr *User) numCommand() {
	num := usr.server.GetUsrNum()
	usr.SendMsgToClient("当前在线用户有: " + strconv.Itoa(num) + " 个.\n")
}

// who指令，查询当前在线用户列表
func (usr *User) whoCommand() {
	usrList, num := usr.server.GetUsrList()
	usr.SendMsgToClient("当前在线用户有: " + strconv.Itoa(num) + " 个,包含以下用户:\n")
	msg := ""
	for k, usrName := range usrList {
		msg += "[" + usrName + "]" + "; "
		// 每行5个输出
		if (k+1)%5 == 0 || k == num-1 {
			usr.SendMsgToClient(msg + "\n")
		}
		k++
	}
}

// 用户命名限制
func (usr *User) nameLimit(usrName string) bool {
	// 不能有空格,大于3字符,小于20字符
	if strings.Contains(usrName, " ") {
		return false
	}
	if len(usrName) >= 20 || len(usrName) <= 3 {
		return false
	}
	return true
}

// rename指令，重命名
func (usr *User) renameCommand() {
	usr.SendMsgToClient("请输入新用户名(不能有空格,大于3字符,小于20字符): ")

	// 读取用户输入
	buf := make([]byte, 20)
	n, err := usr.conn_u.Read(buf)

	// 有错误，且错误不为EOF结束符
	if err != nil && err != io.EOF {
		fmt.Println("Conn Read has err: ", err) // server打印err
		usr.SendMsgToClient("Conn Read has err,请检查客户端是否存在问题.\n")
		return
	}

	// 内容为空直接返回
	if n == 0 {
		return
	}

	// 读取n-1个字符，不读取最后的'\n'
	newName := string(buf[:n-1])

	// 文件名限制
	allow := usr.nameLimit(newName)
	if !allow {
		usr.SendMsgToClient("用户名不符合规范,请重新使用rename.\n")
		return
	}

	// 判断newName是否存在
	usr.server.mapLock.Lock()
	_, ok := usr.server.OnlineMap[newName]
	if ok {
		usr.SendMsgToClient("用户名已存在,请重新使用rename.\n")
	} else {
		delete(usr.server.OnlineMap, usr.Name)
		usr.server.OnlineMap[newName] = usr
		usr.Name = newName
		usr.SendMsgToClient("用户名已更新.\n")
	}
	usr.server.mapLock.Unlock()
}

// 用户消息业务
func (usr *User) DoMsg(msg string) int {
	// 消息处理
	switch msg {
	case "im -exit":
		// 下线，退出命令行
		return -1
	case "im -who":
		// 查询当前在线用户列表
		usr.whoCommand()
	case "im -num":
		// 查询当前在线用户人数
		usr.numCommand()
	case "im -rename":
		// 重命名
		usr.renameCommand()
	default:
		// 调用服务器广播接口
		usr.server.BroadcastUsrMsg(usr, msg)
	}
	return 0
}
