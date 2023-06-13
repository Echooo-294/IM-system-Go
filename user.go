package main

import (
	"net"
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
	for {
		msg := <-usr.Chan_u
		usr.conn_u.Write([]byte(msg + "\n"))
	}
}

// 用户上线
func (usr *User) Online() {
	usr.server.RegisterUsr(usr)
	usr.server.BroadcastUsrMsg(usr, "is Online!")
}

// 用户下线
func (usr *User) Offline() {
	usr.server.DeleteUsr(usr)
	usr.server.BroadcastUsrMsg(usr, "is Offline~")
}

// 用户消息业务
func (usr *User) DoMsg(msg string) {
	// 消息处理

	// 调用服务器广播接口
	usr.server.BroadcastUsrMsg(usr, msg)
}
