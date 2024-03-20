package handler

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"time"

	tls "github.com/refraction-networking/utls"

	"github.com/gorilla/websocket"
	"github.com/nicennnnnnnlee/freedomGo/extend"
	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	"github.com/nicennnnnnnlee/freedomGo/utils/geo"
)

func HandleWsReal(conn net.Conn, conf *config.Local) {
	defer conn.Close()
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

	// 先判断是否直连
	if conf.GeoDomain != nil {
		r := geo.IsDirect(host)
		if (r == nil && conf.GeoDomain.DirectIfNotInRules) ||
			(r != nil && *r) {
			// log.Printf("直连 %s: %s\n", host, port)
			conn2server := getDirectConn(host, port, conf)
			if head == "CONNECT" {
				io.WriteString(conn, HttpsProxyEstablished)
				// conn.Write([]byte(HttpsProxyEstablished))
			} else {
				io.WriteString(conn2server, header)
				// log.Println(header)
				// conn2server.Write([]byte(header))
			}
			go utils.Pip(conn, conn2server)
			utils.Pip(conn2server, conn)
		}
	}
	// 连接远程服务器
	remoteAddr := fmt.Sprintf("%s:%d", conf.RemoteHost, conf.RemotePort)
	remoteHostAddr := fmt.Sprintf("%s:%d", conf.HttpDomain, conf.RemotePort)
	tlsCfg := &tls.Config{
		InsecureSkipVerify: conf.AllowInsecure,
		ServerName:         conf.HttpDomain,
		NextProtos:         []string{"http/1.1"}, // "http/1.1",
		VerifyConnection: func(connState tls.ConnectionState) error {
			if conf.AllowInsecure {
				return nil
			}
			return connState.PeerCertificates[0].VerifyHostname(conf.HttpDomain)
		},
	}
	NetDialContext := func(ctx context.Context, network string, _addr string) (net.Conn, error) {
		var dialer net.Dialer
		dialer.Timeout = time.Second * 5
		if extend.DnsResolver != "" {
			dialer.Resolver = &net.Resolver{
				PreferGo: true,
				Dial: func(ctx0 context.Context, network, address string) (net.Conn, error) {
					return dialer.DialContext(ctx0, "udp", extend.DnsResolver)
				},
			}
		}
		return dialer.DialContext(ctx, network, remoteAddr)
	}
	NetDialTLSContext := func(ctx context.Context, network, _addr string) (net.Conn, error) {
		tcpConn, err := NetDialContext(ctx, network, _addr)
		if err != nil {
			return tcpConn, err
		}
		tlsConn := tls.UClient(tcpConn, tlsCfg, tls.HelloRandomized)
		return tlsConn, err
	}
	var url string
	if conf.RemoteSSL {
		url = "wss://" + remoteHostAddr + conf.HttpPath
	} else {
		url = "ws://" + remoteHostAddr + conf.HttpPath
	}
	// 获取websocket dialer, 并连接
	var dialer = &websocket.Dialer{
		HandshakeTimeout:  45 * time.Second,
		NetDialContext:    NetDialContext,
		NetDialTLSContext: NetDialTLSContext,
	}
	var reqHeader = http.Header{}
	reqHeader.Set("Cookie", GenCookie(conf, host, port))
	reqHeader.Set("User-Agent", conf.HttpUserAgent)
	wsConn, _, err := dialer.Dial(url, reqHeader)
	if err != nil {
		panic(err)
	}
	defer wsConn.Close()

	if head == "CONNECT" {
		io.WriteString(conn, HttpsProxyEstablished)
	} else {
		wsConn.WriteMessage(websocket.BinaryMessage, []byte(header))
	}
	go netConn2WsConn(conn, wsConn)
	wsConn2NetConn(wsConn, conn)
}

func netConn2WsConn(r net.Conn, c *websocket.Conn) {
	defer c.Close()
	defer r.Close()
	defer utils.HandleError()
	buffer := make([]byte, 1024)
	for {
		len, err := r.Read(buffer)
		// log.Printf("netConn2WsConn %d\n", len)
		if len > 0 {
			err = c.WriteMessage(websocket.BinaryMessage, buffer[:len])
			if err != nil {
				panic(err)
			}
		}
		if err != nil {
			panic(err)
		}
	}
}
func wsConn2NetConn(c *websocket.Conn, r net.Conn) {
	defer c.Close()
	defer r.Close()
	defer utils.HandleError()
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			// panic(err)
			break
		}
		// log.Printf("wsConn2NetConn %d\n", len(msg))
		if mt == websocket.BinaryMessage {
			_, err = r.Write(msg)
			if err != nil {
				// panic(err)
				break
			}
		}
	}
}
