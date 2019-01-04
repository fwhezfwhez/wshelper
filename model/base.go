package model

import "time"

type SendOne struct {
	From    string
	To      string
	SendAt  time.Time
	Message string
	Extra   []byte
}

type Reply struct {
	ReplyType  int
	Desc       string
	Tip        string
	Debug      string
	Notice     string
	ReplyValue interface{}
}


