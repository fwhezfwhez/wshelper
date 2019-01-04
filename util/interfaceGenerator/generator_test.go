package interfaceGenerator

import (
	"fmt"
	"testing"
)

func TestGenerator(t *testing.T) {
	rs := ModelToInterface(`
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
   `)
	fmt.Println(rs)
}


// not correct.
func TestCutAfterN(t *testing.T) {
	var src = "// this is a test note of cuteN."
	fmt.Println(len(src))
	fmt.Println(FormatAnnotationByN(src, 5))
}
