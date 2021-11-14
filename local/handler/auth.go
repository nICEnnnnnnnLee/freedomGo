package handler

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
)

var (
	randomWSKey = utils.RandKey()
	authHeaders = "GET %s HTTP/1.1\r\n" +
		"Host: %s\r\n" +
		"User-Agent: %s\r\n" +
		"Accept: */*\r\n" +
		"Accept-Language: zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"Sec-WebSocket-Extensions: permessage-deflate\r\n" +
		"Sec-WebSocket-Key: %s\r\n" +
		"Connection: keep-alive, Upgrade\r\n" +
		"Pragma: no-cache\r\n" +
		"Cache-Control: no-cache\r\n" +
		"Upgrade: websocket\r\n" +
		"Cookie: %s\r\n\r\n"
)

func genHeader(conf *config.Local, domain string, port string) string {
	format := "my_type=1; my_domain=%s; my_port=%s; my_username=%s; my_time=%s; my_token=%s"
	timeNow := strconv.FormatInt(time.Now().UnixMilli(), 10)
	h := md5.New()
	io.WriteString(h, conf.Password)
	io.WriteString(h, conf.Salt)
	io.WriteString(h, timeNow)
	token := fmt.Sprintf("%x", h.Sum(nil))
	cookies := fmt.Sprintf(format, domain, port, conf.Username, timeNow, token)
	var host string
	if conf.RemotePort == 443 {
		host = conf.HttpDomain
	} else {
		host = fmt.Sprintf("%s:%d", conf.HttpDomain, conf.RemotePort)
	}
	headers := fmt.Sprintf(authHeaders, conf.HttpPath, host, conf.HttpUserAgent, randomWSKey, cookies)
	return headers
}

// 和Remote服务器建立连接，并进行鉴权处理
func GetAuthorizedConn(host string, port string, conf *config.Local) net.Conn {
	remoteAddr := fmt.Sprintf("%s:%d", conf.RemoteHost, conf.RemotePort)
	conn2server, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		panic(err)
	}
	// log.Println("remote TCP link established...")
	if conf.RemoteSSL {
		conn2server = tls.Client(conn2server, &tls.Config{
			InsecureSkipVerify: conf.AllowInsecure,
			ServerName:         conf.HttpDomain,
			VerifyConnection: func(connState tls.ConnectionState) error {
				if conf.AllowInsecure {
					return nil
				}
				return connState.PeerCertificates[0].VerifyHostname(conf.HttpDomain)
			},
		})
	}
	// log.Println("remote TLS established...")
	authSend := genHeader(conf, host, port)
	io.WriteString(conn2server, authSend)
	// log.Println("authSend: ", authSend)
	authRecv, err := utils.ReadHeader(conn2server)
	// log.Println("authRecv: ", authRecv)
	if err != nil || !strings.Contains(authRecv, "auth: ok") {
		conn2server.Close()
		panic(utils.ErrAuthNotRight)
	}
	// log.Println("auth success")
	return conn2server
}
