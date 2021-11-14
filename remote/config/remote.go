package config

import (
	"bytes"
	"fmt"
	"reflect"
)

type Remote struct {
	BindHost      string            `yaml:"BindHost"`
	BindPort      uint16            `yaml:"BindPort"`
	UseSSL        bool              `yaml:"UseSSL"`
	SNI           string            `yaml:"SNI"`
	CertPath      string            `yaml:"CertPath"`
	KeyPath       string            `yaml:"KeyPath"`
	Salt          string            `yaml:"Salt"`
	Users         map[string]string `yaml:"Users"`
	ValidHttpPath string            `yaml:"HttpPath"`
}

// type User [2]string

// func NewRemote() *Remote {
// 	users := make(map[string]string)
// 	users["user1"] = "pwd1"
// 	users["user2"] = "pwd2"
// 	return &Remote{
// 		BindPort: 3789,
// 		Salt:     "salt",
// 		Users:    users,
// 	}
// }

func (remote *Remote) String() string {
	// 如果为空，直接返回
	if remote == nil {
		return "<nil>"
	}
	typ := reflect.TypeOf(remote).Elem()
	obj := reflect.ValueOf(remote).Elem()
	numField := typ.NumField()
	buffer := bytes.NewBufferString("\n")
	for i := 0; i < numField; i++ {
		key := typ.Field(i).Name
		value := obj.Field(i)
		fmt.Fprintf(buffer, "%v:\t%v\n", key, value)
		// fmt.Println(key, value)
	}
	return buffer.String()
}
