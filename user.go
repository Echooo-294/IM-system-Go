package main

import (
	"net"
)

type User struct {
	Name    string
	Address string
	Chan_u  chan string
	conn_u  net.Conn
}

// 创建一个user
func NewUser(conn net.Conn) *User {
	// 获取当前地址
	useraddr := conn.RemoteAddr().String()

	usr := &User{
		Name:    useraddr,
		Address: useraddr,
		Chan_u:  make(chan string),
		conn_u:  conn,
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
