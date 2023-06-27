/*
服务端的用户对象
*/
package main

import (
	"net"
	"strconv"
)

// 用户结构体
type User struct {
	Name    string
	Address string
	UsrChan chan string
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
		UsrChan: make(chan string),
		conn:    conn,
		server:  server,
	}

	// 启动监听当前user channel的goroutine
	go usr.ListenUsrMsg()

	return usr
}

// 监听当前user channel，有消息就发给客户端
func (u *User) ListenUsrMsg() {
	// 持续阻塞至用户通道关闭
	for msg := range u.UsrChan {
		u.SendMsgToClient(msg)
	}
	u.SendMsgToClient("UsrChan is Closed.")
}

// 给当前用户的客户端发送消息
func (u *User) SendMsgToClient(msg string) {
	u.conn.Write([]byte(msg + "\n"))
}

// 用户上线
func (u *User) Online() {
	u.server.RegUsr(u)
	u.server.BCUsrMsg(u, "is Online!")
}

// 用户析构关闭资源,统一在server的Handler中defer析构
func (u *User) Close() {
	u.server.DelUsr(u)
	close(u.UsrChan)
	u.conn.Close()
}

// 用户下线
func (u *User) Offline() {
	u.server.BCUsrMsg(u, "is Offline~")
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

// rename指令，重命名
func (u *User) renameCommand(msg string) {
	newName := msg[7:]

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
func (u *User) privateChat(msg string) {
	// 读取用户输入
	tName := msg[3:]

	// 判断是否存在该用户
	u.server.mapLock.Lock()
	targetUsr, isExist := u.server.OnlineMap[tName]
	u.server.mapLock.Unlock()
	if !isExist {
		u.SendMsgToClient("用户名不存在,请重新尝试.")
		return
	}

	targetUsr.SendMsgToClient("[" + u.Name + "] send msg to you : " + msg)
}

// 用户消息业务
func (u *User) DoMsg(msg string) int {
	// 消息处理
	switch msg {
	case "exit":
		// 下线
		return -1
	case "who":
		// 查询当前在线用户列表
		u.whoCommand()
	case "num":
		// 查询当前在线用户人数
		u.numCommand()
	default:
		if len(msg) > 3 && msg[:3] == "to-" {
			// 私聊
			u.privateChat(msg)
		} else if len(msg) > 7 && msg[:7] == "rename-" {
			// 重命名
			u.renameCommand(msg)
		} else {
			// 调用服务器广播接口
			u.server.BCUsrMsg(u, msg)
		}
	}
	return 0
}
