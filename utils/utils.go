package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"runtime/debug"
	"strings"
)

type Connection interface {
	Read(b []byte) (n int, err error)

	Write(b []byte) (n int, err error)

	Close() error
}

var (
	ErrAuthHeaderNotRight = errors.New("remote: auth header is not right")
	ErrHeaderNotRight     = errors.New("header format is not valid")
	ErrAuthNotRight       = errors.New("local: auth is not right")
)

func RandKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
func HandleError() {
	err := recover()
	if err == ErrAuthNotRight || err == ErrHeaderNotRight || err == ErrAuthHeaderNotRight {
		log.Println(err)
		return
	}
	if err != nil && err != io.EOF {
		switch err.(type) {
		case *net.OpError:
			return
		default:
			log.Println(err)
			log.Println(string(debug.Stack()))
			// panic(err)
		}
	}
}

func Pip(from Connection, to Connection) {
	defer from.Close()
	defer to.Close()
	defer HandleError()
	buffer := make([]byte, 1024)
	for {
		len, err := from.Read(buffer)
		if len > 0 {
			// log.Println("pip count: ", from.RemoteAddr(), to.RemoteAddr(), len)
			to.Write(buffer[:len])
			// fmt.Println(string(buffer[:len]))
		}
		if err != nil {
			panic(err)
		}
	}
}


func ReadHeader(conn net.Conn) (string, error) {
	var result []byte = make([]byte, 0)
	buffer := make([]byte, 1024)
	for len(result) < 1024*5 {
		size, err := conn.Read(buffer)
		if err != nil {
			// res := string(result)
			// log.Println("utils.ReadHeader: ", res)
			return "", err
		}
		result = append(result, buffer[0:size]...)
		str := string(result)
		if strings.HasSuffix(str, "\r\n\r\n") {
			return str, nil
		}
	}
	return "", ErrHeaderNotRight
}
