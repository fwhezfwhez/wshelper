package _json

import "time"

type SendOne struct {
	From    string    `json:"from"`
	To      string    `json:"to"`
	SendAt  time.Time `json:"send_at"`
	Message string    `json:"message"`
	Extra   []byte    `json:"extra"`
}

type Reply struct {
	ReplyType  int         `json:"reply_type"` // reply_type range in config/const/const.go
	Desc       string      `json:"desc"`       // describe what this type of reply used to do
	Tip        string      `json:"tip"`        // for REPLY_TIPS 正常提示
	Debug      string      `json:"debug"`      // for REPLY_DEBUG 调试信息
	Notice     string      `json:"notice"`     // for REPLY_NOTICE 系统公告
	ReplyValue interface{} `json:"reply_value"`
}
