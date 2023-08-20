补充服务器控制台快捷命令:设置主动关闭服务器的方式，并通知，关闭后屏显
两个usr查询功能都是在锁恢复后才向客户端输出，应该影响不大
其他机器连接尝试

输入长度超限需要优化，会莫名其妙多一个输入内容为空
请输入新用户名(不能有空格,大于3字符,小于20字符): 
123456789012345678901
输入内容长度不符合要求.
输入内容不得为空.

合理利用context关闭协程,避免forceoffline多次关闭usrconn;
超时强踢会报err：Conn Read has err(ReceiveUsrMsg)多次关闭usr资源
主协程子协程关闭次序不确定

先关闭客户端会报err：对等实体关闭连接
Conn Read has err(ReceiveUsrMsg):  read tcp 127.0.0.1:8888->127.0.0.1:54302: read: connection reset by peer

如何确认已经改好用户名和私聊消息发送成功

手写iocopy会导致客户端输出混乱
	// buf := make([]byte, public.ServeMsgLenLimit+1)
	// for {
	// 	n, err := cli.conn.Read(buf)

	// 	// 有错误，且错误不为EOF结束符
	// 	if err != nil && err != io.EOF {
	// 		fmt.Println("Conn Read has err(DealResponse): ", err) // client打印err
	// 		return
	// 	}

	// 	msg := string(buf)

	// 	// 服务器消息不为空
	// 	if n != 0 {
	// 		fmt.Println(msg)
	// 	}

	// 	if cli.mode == 3 && msg == "用户名已更新." {
	// 		cli.Name = temp
	// 		print(cli.Name)
	// 	}
	// }

CPU好像又异常升高,开了客户端CPU变20+一段时间;似乎又没有这个问题,待测试
没办法发送带空格的消息 可能要使用bufio包中的带缓冲的reader和os包的os.stdin