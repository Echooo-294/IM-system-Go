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
	conn    net.Conn
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
		conn:    conn,
		server:  server,
	}

	// 启动监听当前user channel的goroutine
	go usr.ListenUsrMsg()

	return usr
}

// 监听当前user channel，有消息就发给Usr
func (u *User) ListenUsrMsg() {
	for msg := range u.ChanUsr {
		u.SendMsgToClient(msg)
	}
}

// 给当前用户的客户端发送消息
func (u *User) SendMsgToClient(msg string) {
	u.conn.Write([]byte(msg + "\n"))
}

// 用户上线
func (u *User) Online() {
	u.server.RegisterUsr(u)
	u.server.BroadcastUsrMsg(u, "is Online!")
}

// 用户析构关闭资源,统一在server的Handler中defer析构
func (u *User) CloseResources() {
	u.server.DeleteUsr(u)
	close(u.ChanUsr)
	u.conn.Close()
}

// 用户下线
func (u *User) Offline() {
	u.server.BroadcastUsrMsg(u, "is Offline~")
	u.SendMsgToClient("You are Offline.")
}

// 用户被强制踢出
func (u *User) ForceOffline() {
	u.SendMsgToClient("You are ForceOffline.")
}

// num指令，查询当前在线用户人数
func (u *User) numCommand() {
	num := u.server.UsrNum()
	u.SendMsgToClient("当前在线用户有: " + strconv.Itoa(num) + " 个.")
}

// who指令，查询当前在线用户列表
func (u *User) whoCommand() {
	uList, num := u.server.UsrList()
	u.SendMsgToClient("当前在线用户有: " + strconv.Itoa(num) + " 个,包含以下用户:")
	msg := ""
	for k, uName := range uList {
		msg += "[" + uName + "]" + "; "
		// 每行5个输出
		if (k+1)%5 == 0 || k == num-1 {
			u.SendMsgToClient(msg)
		}
		k++
	}
}

// 用户命名限制
func (u *User) nameLimit(uName string) bool {
	// 不能有空格
	return !strings.Contains(uName, " ")
}

// 读取用户输入
func (u *User) readUsr(tips string, limitDn int, limitUp int) (string, bool) {
	u.SendMsgToClient(tips)
	buf := make([]byte, limitUp+1)
	n, err := u.conn.Read(buf)

	// 有错误，且错误不为EOF结束符
	if err != nil && err != io.EOF {
		fmt.Println("Conn Read has err(readUsr): ", err) // server打印err
		return "", false
	}

	// 消息为空（退出该步）
	if n == 0 {
		return "", false
	}

	// 输入不得仅有换行符
	if string(buf[0]) == "\n" {
		u.SendMsgToClient("输入内容不得为空.")
		return "", false
	}

	// 内容长度限制
	if n >= limitUp || n <= limitDn {
		u.SendMsgToClient("输入内容长度不符合要求.")
		return "", false
	}

	// 读取n-1个字符，不读取最后的'\n'
	return string(buf[:n-1]), true
}

// rename指令，重命名
func (u *User) renameCommand() {
	// 读取n-1个字符，不读取最后的'\n'
	newName, ok := u.readUsr("请输入新用户名(不能有空格,大于3字符,小于20字符): ", 3, 20)
	if !ok {
		return
	}

	// 不得与当前用户名相同
	if newName == u.Name {
		u.SendMsgToClient("用户名不得与当前用户名相同,请重新尝试.")
		return
	}

	// 用户名限制
	allow := u.nameLimit(newName)
	if !allow {
		u.SendMsgToClient("用户名不符合规范,请重新尝试.")
		return
	}

	// 判断newName是否存在
	u.server.mapLock.Lock()
	_, isExist := u.server.OnlineMap[newName]
	if isExist {
		u.SendMsgToClient("用户名已存在,请重新尝试.")
	} else {
		delete(u.server.OnlineMap, u.Name)
		u.server.OnlineMap[newName] = u
		u.Name = newName
		u.SendMsgToClient("用户名已更新.")
	}
	u.server.mapLock.Unlock()
}

// 私聊功能
func (u *User) privateChat() {
	// 读取用户输入
	tName, ok1 := u.readUsr("请输入私聊对象的用户名: ", 3, 20)
	if !ok1 {
		return
	}

	// 判断是否是给自己发送
	if tName == u.Name {
		u.SendMsgToClient("不得与自己聊天,请重新尝试.")
		return
	}

	// 判断是否存在该用户
	u.server.mapLock.Lock()
	targetUsr, isExist := u.server.OnlineMap[tName]
	u.server.mapLock.Unlock()
	if !isExist {
		u.SendMsgToClient("用户名不存在,请重新尝试.")
		return
	}

	// 向该用户发送消息
	msg, ok2 := u.readUsr("请输入要发送的内容: ", 1, usrMsgLenLimit)
	if !ok2 {
		return
	}
	targetUsr.SendMsgToClient("[" + u.Name + "] send msg to you : " + msg)
}

// 用户消息业务
func (u *User) DoMsg(msg string) int {
	// 消息处理
	switch msg {
	case "im -exit":
		// 下线
		return -1
	case "im -who":
		// 查询当前在线用户列表
		u.whoCommand()
	case "im -num":
		// 查询当前在线用户人数
		u.numCommand()
	case "im -rename":
		// 重命名
		u.renameCommand()
	case "im -to":
		// 私聊
		u.privateChat()
	default:
		// 调用服务器广播接口
		u.server.BroadcastUsrMsg(u, msg)
	}
	return 0
}
