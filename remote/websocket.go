package remote

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/nicennnnnnnlee/freedomGo/remote/config"
	"github.com/nicennnnnnnlee/freedomGo/remote/handler"
	"github.com/nicennnnnnnlee/freedomGo/utils"
)

const (
	response_403 = "HTTP/1.1 403 Forbidden\r\n" + "Content-Length: 0\r\n" + "Connection: closed\r\n\r\n"
	response_101 = "HTTP/1.1 101 Switching Protocols\r\n" + "auth: ok\r\n" + "Sec-WebSocket-Accept: %s" +
		"\r\nUpgrade: websocket\r\n" + "Connection: Upgrade\r\n\r\n"
)

func Make101Response(conn net.Conn) {
	fmt.Fprintf(conn, response_101, utils.RandKey())
	// io.WriteString(conn, response_101)
}

func Make403Response(conn net.Conn) {
	io.WriteString(conn, response_403)
}
func handleClient(conn net.Conn, conf *config.Remote, tlsConf *tls.Config) {
	defer conn.Close()
	defer utils.HandleError()
	// log.Println("tcp conn established...")
	if conf.UseSSL {
		conn = tls.Server(conn, tlsConf)
	}

	authRecv, err := utils.ReadHeader(conn)
	if err != nil {
		panic(utils.ErrAuthHeaderNotRight)
	}
	if conf.ValidHttpPath != "" && !strings.HasPrefix(authRecv, "GET "+conf.ValidHttpPath+" HTTP/1.1") {
		Make403Response(conn)
		return
	}
	// log.Println("auth header received...")
	remoteAddr := handler.GetRemoteAddr(authRecv, conf)
	if remoteAddr == nil {
		Make403Response(conn)
	} else {
		Make101Response(conn)
		conn2server := handler.GetRemoteConn(remoteAddr, conf)
		go utils.Pip(conn, conn2server)
		utils.Pip(conn2server, conn)
	}

}

func StartWs(conf *config.Remote) {
	var tlsConfig *tls.Config
	if conf.UseSSL {
		tlsCert := &utils.TlsCert{
			CertPath:       conf.CertPath,
			KeyPath:        conf.KeyPath,
			AttempDuration: time.Minute * 5,
		}
		GetCertFunc := tlsCert.GetCertFunc()
		tlsConfig = &tls.Config{
			ServerName:     conf.SNI,
			GetCertificate: GetCertFunc,
		}
	}

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
			go handleClient(conn, conf, tlsConfig)
		}
	}
}
