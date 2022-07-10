package handler

import (
	"fmt"
	"io"
	"net"
	"regexp"

	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	"github.com/nicennnnnnnlee/freedomGo/utils/geo"
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

	conn2server := getRightConn(host, port, conf)
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

func getRightConn(host string, port string, conf *config.Local) net.Conn {
	if conf.GeoDomain != nil {
		r := geo.IsDirect(host)
		if (r == nil && conf.GeoDomain.DirectIfNotInRules) ||
			(r != nil && *r) {
			// log.Printf("直连 %s: %s\n", host, port)
			return getDirectConn(host, port, conf)
		}
	}
	// log.Printf("走代理 %s: %s\n", host, port)
	return GetAuthorizedConn(host, port, conf)
}

func getDirectConn(host string, port string, conf *config.Local) net.Conn {
	remoteAddr := fmt.Sprintf("%s:%s", host, port)
	// conn2server, err := net.Dial("tcp", remoteAddr)
	conn2server, err := utils.DialTCP(remoteAddr, conf.DNSServer)
	if err != nil {
		panic(err)
	}
	return conn2server
}
