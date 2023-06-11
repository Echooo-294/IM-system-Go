package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 用户信息表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 服务器消息广播通道
	Chan_s chan string
}

// 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Chan_s:    make(chan string),
	}
	return server
}

// 启动服务器
func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// defer listen socket
	defer listener.Close()

	// 启动监听广播通道的goroutine
	go server.ListenServerMsg()

	// 持续接受链接
	for {
		conn, err := listener.Accept()
		if err == nil {
			go server.Handler(conn)
		} else {
			fmt.Println("listener accept err:", err)
		}
	}
}

// 广播通道监听
func (server *Server) ListenServerMsg() {
	for {
		msg := <-server.Chan_s

		// 广播
		server.mapLock.Lock()
		for _, usr := range server.OnlineMap {
			usr.Chan_u <- msg
		}
		server.mapLock.Unlock()
	}
}

// 用户上线的消息写入广播通道
func (server *Server) Broadcast_usrMsg(usr *User, msg string) {
	usrMsg := "The user with the name " + usr.Name + msg
	server.Chan_s <- usrMsg
}

// 业务处理
func (server *Server) Handler(conn net.Conn) {
	// 接收链接后新建一个usr
	usr := NewUser(conn)

	// 登记用户信息
	server.mapLock.Lock()
	server.OnlineMap[usr.Name] = usr
	server.mapLock.Unlock()

	// 对用户上线消息进行广播
	server.Broadcast_usrMsg(usr, " is Online.")

	// 阻塞避免主协程结束
	select {}
}
