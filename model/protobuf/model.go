package _protobuf

import "time"

type SendOne struct {
	From    string
	To      string
	SendAt  time.Time
	Message string
	Extra   []byte
}

type Reply struct {
	ReplyType  int    // reply_type range in config/const/const.go
	Desc       string // describe what this type of reply used to do
	Tip        string // for REPLY_TIPS 正常提示
	Debug      string // for REPLY_DEBUG 调试信息
	Notice     string // for REPLY_NOTICE 系统公告
	ReplyValue interface{}
}
