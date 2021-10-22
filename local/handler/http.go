package handler

import (
	"freedomGo/config"
	"freedomGo/utils"
	"io"
	"net"
	"regexp"
)

const (
	HttpsProxyEstablished = "HTTP/1.1 200 Connection Established\r\n\r\n"
)

func HandleHttp(conn net.Conn, conf *config.Local) {
	header, err := utils.ReadHeader(conn)
	if err != nil {
		panic(err)
	}
	reg := regexp.MustCompile(`(CONNECT|Host:) ([^ :\r\n]+)(?::(\d+))?`)
	matches := reg.FindStringSubmatch(header)
	if matches == nil {
		panic(utils.ErrHeaderNotRight)
	}
	head, host, port := matches[1], matches[2], matches[3]
	if port == "" {
		port = "80"
	}

	conn2server := GetAuthorizedConn(host, port, conf)
	if head == "CONNECT" {
		io.WriteString(conn, HttpsProxyEstablished)
		// conn.Write([]byte(HttpsProxyEstablished))
	} else {
		io.WriteString(conn2server, header)
		// log.Println(header)
		conn2server.Write([]byte(header))

	}
	go utils.Pip(conn, conn2server)
	utils.Pip(conn2server, conn)
	// time.Sleep(time.Second * 60)
}
