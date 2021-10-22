package local

import (
	"fmt"
	"freedomGo/config"
	"freedomGo/local/handler"
	"freedomGo/utils"
	"net"
)

func handleClient(conn net.Conn, conf *config.Local) {
	defer conn.Close()
	defer utils.HandleError()
	handler.HandleHttp(conn, conf)
}

func Start(conf *config.Local) {

	fmt.Println("服务器开始监听...")
	addr := fmt.Sprintf("%s:%d", conf.BindHost, conf.BindPort)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Accept err:", err)
		} else {
			go handleClient(conn, conf)
		}
	}
}
