/*
服务器实现,tcp协议
*/
package main

import (
	"fmt"
	"im_system/public"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

// 服务器结构体
type Server struct {
	Ip   string
	Port int
	// 用户信息表及其读写锁
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 服务器消息广播通道
	ServeChan chan string
}

// 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		ServeChan: make(chan string),
	}
	return server
}

// 启动服务器
func (s *Server) Start() {
	fmt.Println("Our Server is RUNNING!!!")

	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	// defer listen socket
	defer listener.Close()

	// 启动监听广播通道的goroutine
	go s.ListenServeMsg()

	// 持续接受链接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err: ", err)
		} else {
			go s.Handler(conn)
		}
	}
}

// 广播通道监听
func (s *Server) ListenServeMsg() {
	// 持续阻塞至广播通道关闭
	for msg := range s.ServeChan {
		// 有消息则广播
		s.mapLock.Lock()
		for _, usr := range s.OnlineMap {
			usr.UsrChan <- msg
		}
		s.mapLock.Unlock()
	}
	fmt.Println("ServeChan is Closed.")
}

// 将服务器的消息写入广播通道
func (s *Server) BCServeMsg(msg string) {
	serverMsg := "~ [Server] : " + msg
	s.ServeChan <- serverMsg
}

// 将某用户的消息写入广播通道
func (s *Server) BCUsrMsg(usr *User, msg string) {
	usrMsg := "# [" + usr.Name + "] : " + msg
	s.ServeChan <- usrMsg
}

// 持续接收用户输入的消息进行处理
func (s *Server) RecvUsrMsg(usr *User, conn net.Conn, isLive chan bool) {
	buf := make([]byte, public.UsrMsgMaxLen+1)
	for {
		n, err := conn.Read(buf)

		// 有错误，且错误不为EOF结束符
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read has err(ReceiveUsrMsg): ", err) // server打印err
			return
		}

		// 消息为空（退出程序）表示下线
		if n == 0 {
			usr.Offline()
			isLive <- false
			return
		}

		// 仅输入回车,则等待重新输入
		if string(buf[0]) == "\n" {
			usr.SendMsgToClient("输入内容不得为空.")
			continue
		}

		// 输入内容长度超限
		if n >= public.UsrMsgMaxLen {
			usr.SendMsgToClient("输入内容长度不符合要求.")
			continue
		}

		// 输入不合要求也会导致用户成为不活跃状态

		// 读取n-1个字符，不读取最后的'\n'，进行用户消息业务
		msg := string(buf[:n-1])
		status := usr.DoMsg(msg)

		// 对用户消息业务返回的status进行处理
		switch status {
		case -1:
			// 用户输入exit,退出处理该用户消息的协程
			usr.Offline()
			isLive <- false
			return
		default:
			// 用户的非退出指令的消息表示ta是活跃的
			isLive <- true
		}
	}
}

// 登记用户信息
func (s *Server) RegUsr(usr *User) {
	s.mapLock.Lock()
	s.OnlineMap[usr.Name] = usr
	s.mapLock.Unlock()
}

// 删除用户信息
func (s *Server) DelUsr(usr *User) {
	s.mapLock.Lock()
	delete(s.OnlineMap, usr.Name)
	s.mapLock.Unlock()
}

// 获取服务器当前在线人数
func (s *Server) UsrNum() int {
	s.mapLock.Lock()
	num := len(s.OnlineMap)
	s.mapLock.Unlock()
	return num
}

// 获取服务器当前在线用户列表
func (s *Server) UsrList() ([]string, int) {
	s.mapLock.Lock()
	num := len(s.OnlineMap)
	usrList := make([]string, num)
	i := 0
	for _, usr := range s.OnlineMap {
		usrList[i] = usr.Name
		i++
	}
	s.mapLock.Unlock()
	return usrList, num
}

// 对刚上线的用户进行服务器公告
func (s *Server) OnlineNotice(usr *User) {
	// 提示当前在线人数
	msg := "欢迎来到IM服务器, " + "当前 " + strconv.Itoa(s.UsrNum()) + " 人在线."
	usr.SendMsgToClient(msg)
}

// 业务处理
func (s *Server) Handler(conn net.Conn) {
	// 接收链接后新建一个usr
	usr := NewUser(conn, s)

	// 该函数退出后关闭usr资源
	defer usr.Close()

	// 用户上线，登记并广播
	usr.Online()

	// 对刚上线的用户进行服务器公告
	s.OnlineNotice(usr)

	// 记录用户活跃状态的通道
	isLive := make(chan bool)

	// 持续接收用户输入的消息进行处理
	go s.RecvUsrMsg(usr, conn, isLive)

	// 阻塞
	for {
		select {
		case a := <-isLive:
			// true表示用户活跃,空实现辅助重置下方time通道;false表示用户已下线
			if !a {
				return
			}
		case <-time.After(public.UsrMaxTime):
			// 当前用户不活跃状态超时,强制当前用户下线
			usr.ForceOffline()
			return
		}
	}
}
