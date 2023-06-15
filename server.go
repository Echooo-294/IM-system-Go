package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 用户信息表及其读写锁
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
		usr_len_limit: 128,
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
	for msg := range server.Chan_s {
		// 广播
		server.mapLock.Lock()
		for _, usr := range server.OnlineMap {
			usr.Chan_u <- msg
		}
		server.mapLock.Unlock()
	}
}

// 将服务器的消息写入广播通道
func (server *Server) BroadcastServerMsg(msg string) {
	serverMsg := "The Current Server : " + msg
	server.Chan_s <- serverMsg
}

// 将某用户的消息写入广播通道
func (server *Server) BroadcastUsrMsg(usr *User, msg string) {
	usrMsg := "The user [" + usr.Name + "] : " + msg
	server.Chan_s <- usrMsg
}

// 持续接收用户输入的消息进行处理
func (server *Server) ReceiveUsrMsg(usr *User, conn net.Conn) {
	buf := make([]byte, server.usr_len_limit)
	for {
		n, err := conn.Read(buf)
		// 消息为空（退出程序）表示下线
		if n == 0 {
			usr.Offline()
			return
		}

		// 有错误，且错误不为EOF结束符
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read has err: ", err)
			usr.SendMsgToClient("Conn Read has err,请检查客户端是否存在问题.\n")
			return
		}

		// 读取n-1个字符，不读取最后的'\n'，进行消息广播
		msg := string(buf[:n-1])

		// 用户消息业务
		status := usr.DoMsg(msg)

		if status == -1 {
			// 用户输入exit,退出处理该用户消息的协程
			return
		}
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

// 获取服务器当前在线人数
func (server *Server) GetUsrNum() int {
	server.mapLock.Lock()
	num := len(server.OnlineMap)
	server.mapLock.Unlock()
	return num
}

// 获取服务器当前在线用户列表
func (server *Server) GetUsrList() ([]string, int) {
	server.mapLock.Lock()
	num := len(server.OnlineMap)
	usrList := make([]string, num)
	i := 0
	for _, usr := range server.OnlineMap {
		usrList[i] = usr.Name
		i++
	}
	server.mapLock.Unlock()
	return usrList, num
}

// // 对刚上线的用户进行服务器公告
// func (server *Server) NoticeOnline() {
// 	// 提示当前在线人数
// 	msg := "当前 " + strconv.Itoa(server.GetUsrNum()) + " 人在线."
// 	server.BroadcastServerMsg(msg)
// }

// 业务处理
func (server *Server) Handler(conn net.Conn) {
	// 接收链接后新建一个usr
	usr := NewUser(conn, server)

	// 用户上线，登记并广播
	usr.Online()

	// 对刚上线的用户进行服务器公告
	// server.NoticeOnline()

	// 持续接收用户输入的消息进行处理
	go server.ReceiveUsrMsg(usr, conn)

	// 阻塞避免主协程结束
	select {}
}
