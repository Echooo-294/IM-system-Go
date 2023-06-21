补充服务器控制台快捷命令:设置主动关闭服务器的方式，并通知，关闭后屏显
两个usr查询功能都是在锁恢复后才向客户端输出，应该影响不大
用户上线获取服务器公告
其他机器连接尝试

输入长度超限需要优化，会莫名其妙多一个输入内容为空
请输入新用户名(不能有空格,大于3字符,小于20字符): 
123456789012345678901
输入内容长度不符合要求.
输入内容不得为空.

合理利用context关闭协程,避免forceoffline多次关闭usrconn;也可以改成关闭usr统一在handler或recive中执行
超时强踢会报err：Conn Read has err(ReceiveUsrMsg)多次关闭usr资源
先关闭客户端会报err：对等实体关闭连接
Conn Read has err(ReceiveUsrMsg):  read tcp 127.0.0.1:8888->127.0.0.1:54302: read: connection reset by peer