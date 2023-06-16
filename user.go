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
	ChanUsr chan string
	connUsr net.Conn
	server  *Server
}

// 创建一个user
func NewUser(conn net.Conn, server *Server) *User {
	// 获取当前地址
	useraddr := conn.RemoteAddr().String()

	usr := &User{
		Name:    useraddr,
		Address: useraddr,
		ChanUsr: make(chan string),
		connUsr: conn,
		server:  server,
	}

	// 启动监听当前user channel的goroutine
	go usr.ListenUsrMsg()

	return usr
}

// 监听当前user channel，有消息就发给Client
func (usr *User) ListenUsrMsg() {
	for msg := range usr.ChanUsr {
		usr.SendMsgToClient(msg)
	}
}

// 给当前客户端发送消息
func (usr *User) SendMsgToClient(msg string) {
	usr.connUsr.Write([]byte(msg + "\n"))
}

// 用户上线
func (usr *User) Online() {
	usr.server.RegisterUsr(usr)
	usr.server.BroadcastUsrMsg(usr, "is Online!")
}

// 用户析构关闭资源,统一在server的Handler中defer析构
func (usr *User) CloseResources() {
	usr.server.DeleteUsr(usr)
	close(usr.ChanUsr)
	usr.connUsr.Close()
}

// 用户下线
func (usr *User) Offline() {
	usr.server.BroadcastUsrMsg(usr, "is Offline~")
	usr.SendMsgToClient("You are Offline.")
}

// 用户被强制踢出
func (usr *User) ForceOffline() {
	usr.SendMsgToClient("You are ForceOffline.")
}

// num指令，查询当前在线用户人数
func (usr *User) numCommand() {
	num := usr.server.GetUsrNum()
	usr.SendMsgToClient("当前在线用户有: " + strconv.Itoa(num) + " 个.")
}

// who指令，查询当前在线用户列表
func (usr *User) whoCommand() {
	usrList, num := usr.server.GetUsrList()
	usr.SendMsgToClient("当前在线用户有: " + strconv.Itoa(num) + " 个,包含以下用户:")
	msg := ""
	for k, usrName := range usrList {
		msg += "[" + usrName + "]" + "; "
		// 每行5个输出
		if (k+1)%5 == 0 || k == num-1 {
			usr.SendMsgToClient(msg)
		}
		k++
	}
}

// 用户命名限制
func (usr *User) nameLimit(usrName string) bool {
	// 不能有空格
	return !strings.Contains(usrName, " ")
}

// 读取用户输入
func (usr *User) readClient(tips string, limitDn int, limitUp int) (string, bool) {
	usr.SendMsgToClient(tips)
	buf := make([]byte, limitUp+1)
	n, err := usr.connUsr.Read(buf)

	// 有错误，且错误不为EOF结束符
	if err != nil && err != io.EOF {
		fmt.Println("Conn Read has err(readClient): ", err) // server打印err
		return "", false
	}

	// 消息为空（退出该步）
	if n == 0 {
		return "", false
	}

	// 输入不得仅有换行符
	if string(buf[0]) == "\n" {
		usr.SendMsgToClient("输入内容不得为空.")
		return "", false
	}

	// 内容长度限制
	if n >= limitUp || n <= limitDn {
		usr.SendMsgToClient("输入内容长度不符合要求.")
		return "", false
	}

	// 读取n-1个字符，不读取最后的'\n'
	return string(buf[:n-1]), true
}

// rename指令，重命名
func (usr *User) renameCommand() {
	// 读取n-1个字符，不读取最后的'\n'
	newName, ok := usr.readClient("请输入新用户名(不能有空格,大于3字符,小于20字符): ", 3, 20)
	if !ok {
		return
	}

	// 不得与当前用户名相同
	if newName == usr.Name {
		usr.SendMsgToClient("用户名不得与当前用户名相同,请重新尝试.")
		return
	}

	// 用户名限制
	allow := usr.nameLimit(newName)
	if !allow {
		usr.SendMsgToClient("用户名不符合规范,请重新尝试.")
		return
	}

	// 判断newName是否存在
	usr.server.mapLock.Lock()
	_, isExist := usr.server.OnlineMap[newName]
	if isExist {
		usr.SendMsgToClient("用户名已存在,请重新尝试.")
	} else {
		delete(usr.server.OnlineMap, usr.Name)
		usr.server.OnlineMap[newName] = usr
		usr.Name = newName
		usr.SendMsgToClient("用户名已更新.")
	}
	usr.server.mapLock.Unlock()
}

// 私聊功能
func (usr *User) privateChat() {
	// 读取用户输入
	targetName, ok1 := usr.readClient("请输入私聊对象的用户名: ", 3, 20)
	if !ok1 {
		return
	}

	// 判断是否是给自己发送
	if targetName == usr.Name {
		usr.SendMsgToClient("不得与自己聊天,请重新尝试.")
		return
	}

	// 判断是否存在该用户
	usr.server.mapLock.Lock()
	targetUsr, isExist := usr.server.OnlineMap[targetName]
	usr.server.mapLock.Unlock()
	if !isExist {
		usr.SendMsgToClient("用户名不存在,请重新尝试.")
		return
	}

	// 向该用户发送消息
	msg, ok2 := usr.readClient("请输入要发送的内容: ", 1, usrMsgLenLimit)
	if !ok2 {
		return
	}
	targetUsr.SendMsgToClient("[" + usr.Name + "] send msg to you : " + msg)
}

// 用户消息业务
func (usr *User) DoMsg(msg string) int {
	// 消息处理
	switch msg {
	case "im -exit":
		// 下线
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
	case "im -to":
		// 私聊
		usr.privateChat()
	default:
		// 调用服务器广播接口
		usr.server.BroadcastUsrMsg(usr, msg)
	}
	return 0
}
