package main

import (
	"net"
	"strconv"
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
	//获取当前地址
	useraddr := conn.RemoteAddr().String()

	usr := &User{
		Name:    useraddr,
		Address: useraddr,
		Chan_u:  make(chan string),
		conn_u:  conn,
		server:  server,
	}

	//启动监听当前user channel的goroutine
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

// 用户析构销毁
func (usr *User) destroy() {
	usr.server.DeleteUsr(usr)
	close(usr.Chan_u)
	usr.conn_u.Close()
}

// 用户下线
func (usr *User) Offline() {
	usr.server.BroadcastUsrMsg(usr, "is Offline~")
	usr.destroy()
}

// num指令，查询当前在线用户人数
func (usr *User) numCommand() {
	num := usr.server.GetUsrNum()
	usr.SendMsgToClient("当前在线用户有: " + strconv.Itoa(num) + " 个.\n")
}

// who指令，查询当前在线用户列表
func (usr *User) whoCommand() {
	newMsg := usr.server.GetUsrList()
	usr.SendMsgToClient("当前在线用户有: \n")
	for _, usrName := range newMsg {
		usr.SendMsgToClient("[" + usrName + "]" + "\n")
	}
}

// rename指令，重命名，指令格式为"rename:newName"
func (usr *User) renameCommand(msg string) {
	newName := msg[7:]
	// 判断newName是否存在
	usr.server.mapLock.Lock()
	_, ok := usr.server.OnlineMap[newName]
	if ok {
		usr.SendMsgToClient("用户名已存在,请重新使用rename指令.\n")
	} else {
		delete(usr.server.OnlineMap, usr.Name)
		usr.server.OnlineMap[newName] = usr
		usr.Name = newName
		usr.SendMsgToClient("用户名已更新.\n")
	}
	usr.server.mapLock.Unlock()
}

// 用户消息业务
func (usr *User) DoMsg(msg string) {
	//消息处理
	if msg == "exit" {
		//下线，退出命令行

	} else if msg == "who" {
		//查询当前在线用户列表
		usr.whoCommand()
	} else if msg == "num" {
		//查询当前在线用户人数
		usr.numCommand()
	} else if len(msg) > 7 && msg[:7] == "rename:" {
		//重命名
		usr.renameCommand(msg)
	} else {
		//调用服务器广播接口
		usr.server.BroadcastUsrMsg(usr, msg)
	}

}
