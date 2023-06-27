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
