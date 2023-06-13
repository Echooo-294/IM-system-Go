package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
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

	// 用户消息长度限制
	usr_len_limit int
}

// 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:            ip,
		Port:          port,
		OnlineMap:     make(map[string]*User),
		Chan_s:        make(chan string),
		usr_len_limit: 4096,
	}
	return server
}

// 启动服务器
func (server *Server) Start() {
	fmt.Println("Our Server is RUNNING!!!")

	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	// defer listen socket
	defer listener.Close()

	// 启动监听广播通道的goroutine
	go server.ListenServerMsg()

	// defer fmt.Println("Our Server is CLOSED...")

	// 持续接受链接
	for {
		conn, err := listener.Accept()
		if err == nil {
			go server.Handler(conn)
		} else {
			fmt.Println("listener accept err: ", err)
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

// 服务器进行消息广播
func (server *Server) BroadcastServerMsg(msg string) {
	serverMsg := "The Current Server : " + msg
	server.Chan_s <- serverMsg
}

// 将某用户的消息写入广播通道
func (server *Server) BroadcastUsrMsg(usr *User, msg string) {
	usrMsg := "The user (" + usr.Name + ") : " + msg
	server.Chan_s <- usrMsg
}

// 持续接收用户输入的消息进行处理
func (server *Server) ReceiveUsrMsg(usr *User, conn net.Conn) {
	buf := make([]byte, server.usr_len_limit)
	for {
		// 考虑是否需要在某处结束协程

		n, err := conn.Read(buf)
		// 消息为空（退出程序）表示下线
		if n == 0 {
			usr.Offline()
			return
		}
		// 有错误，且错误不为EOF结束符
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read has err: ", err)
			return
		}

		// 读取n-1个字符，不读取最后的'\n'，进行消息广播
		msg := string(buf[:n-1])

		// 用户消息业务
		usr.DoMsg(msg)
	}
}

// 登记用户信息
func (server *Server) RegisterUsr(usr *User) {
	server.mapLock.Lock()
	server.OnlineMap[usr.Name] = usr
	server.mapLock.Unlock()
}

// 删除用户信息
func (server *Server) DeleteUsr(usr *User) {
	server.mapLock.Lock()
	delete(server.OnlineMap, usr.Name)
	server.mapLock.Unlock()
}

// 业务处理
func (server *Server) Handler(conn net.Conn) {
	// 接收链接后新建一个usr
	usr := NewUser(conn, server)

	// 用户上线，登记并广播
	usr.Online()

	// 服务器广播当前用户人数
	server.BroadcastServerMsg(strconv.Itoa(len(server.OnlineMap)))

	// 持续接收用户输入的消息进行处理
	go server.ReceiveUsrMsg(usr, conn)

	// 阻塞避免主协程结束
	select {}
}
