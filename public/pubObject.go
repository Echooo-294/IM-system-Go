package public

import (
	"time"
)

// 服务器响应消息长度限制
const ServeMsgMaxLen int = 128

// 用户消息长度限制
const UsrMsgMaxLen int = 512

// 用户超时限制
const UsrMaxTime time.Duration = time.Second * 120

// 客户端命令字符串
var UsrOrderList map[string]string

func init() {
	UsrOrderList = make(map[string]string)
	UsrOrderList["exit"] = "im -exit"
	UsrOrderList["who"] = "im -who"
	UsrOrderList["num"] = "im -num"
	UsrOrderList["rename"] = "im -rename "
	UsrOrderList["to"] = "im -to "
}
