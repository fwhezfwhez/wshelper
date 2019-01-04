package util

import (
	"crypto/md5"
	"fmt"
	"strings"
	"testing"
)

// MD5 encrypt
func MD5(rawMsg string) string {
	data := []byte(rawMsg)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has)
	return strings.ToUpper(md5str1)
}

// used for testing
func Assert(con bool, t *testing.T,msg ...interface{}){
	if !con {
		t.Fatal(msg...)
	}
}

// used for testing
func Assertf(con bool, t *testing.T,format string, msg ...interface{}){
	if !con {
		t.Fatal(fmt.Sprintf(format,msg...))
	}
}
