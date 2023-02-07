package handler

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nicennnnnnnlee/freedomGo/extend"
	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	"github.com/nicennnnnnnlee/freedomGo/utils/geo"
	"golang.org/x/net/http2"
)

func HandleHttp2(conn net.Conn, conf *config.Local) {
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
	tlsCfg := &tls.Config{
		InsecureSkipVerify: conf.AllowInsecure,
		ServerName:         conf.HttpDomain,
		NextProtos:         []string{"h2"}, // "http/1.1",
		VerifyConnection: func(connState tls.ConnectionState) error {
			if conf.AllowInsecure {
				return nil
			}
			return connState.PeerCertificates[0].VerifyHostname(conf.HttpDomain)
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var conn2server net.Conn
	var url string
	// conn2server, err := DialTCP(remoteAddr, dnsResolver)
	if conf.RemoteSSL {
		conn2server, err = extend.DialTLSContext(ctx, "tcp", remoteAddr, tlsCfg)
		url = "https://" + conf.HttpDomain + conf.HttpPath
	} else {
		// conn2server, err = extend.DialTCPContext(ctx, remoteAddr)
		// url = "http://" + conf.HttpDomain + conf.HttpPath
		log.Println("http2 mode must set conf.RemoteSSL to true!")
		return
	}
	if err != nil {
		log.Println("connection establish failed: ", remoteAddr)
		return
	}
	defer conn2server.Close()
	// tlsConn := tls.Client(conn2server, tlsCfg)

	// 获取h2Client
	transport := &http2.Transport{
		TLSClientConfig: tlsCfg,
		DialTLSContext:  extend.DialTLSContext,
	}
	h2Client, err := transport.NewClientConn(conn2server)
	defer h2Client.Shutdown(ctx)
	if err != nil {
		log.Println("h2Client initialize failed...")
		return
	}
	// 向服务器发送消息，建立tunnel
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", GenCookie(conf, host, port))
	id := fmt.Sprintf("%s:%s:%d", host, port, time.Now().UnixNano())
	req.Header.Set("ID", id)
	res, err := h2Client.RoundTrip(req)
	if err != nil {
		log.Println("h2Client initialize failed...")
		panic(err)
		// return
	}
	// log.Println("h2Client 发送完第一次请求...")
	if res.Header.Get("auth") == "ok" {

		// 将从local收到的发送给 remote
		go func() {
			defer extend.HandleError()
			defer conn2server.Close()
			defer h2Client.Shutdown(ctx)
			// 回复消息 HttpsProxyEstablished
			if head == "CONNECT" {
				io.WriteString(conn, HttpsProxyEstablished)
				// conn.Write([]byte(HttpsProxyEstablished))
			} else {
				if !h2Client.CanTakeNewRequest() {
					panic(errors.New("h2Client 无法再发送新的request"))
				}
				ctx2, cancel2 := context.WithCancel(ctx)
				req, _ := http.NewRequestWithContext(ctx2, "POST", url, strings.NewReader(header))
				req.Header.Set("ID", id)
				req.Header.Set("Next", "1")
				_, err := h2Client.RoundTrip(req)
				if err != nil {
					panic(errors.New("h2Client request发送失败"))
				}
				// h2Client.Shutdown(ctx2)
				cancel2()
			}
			if !h2Client.CanTakeNewRequest() {
				panic(errors.New("h2Client 无法再发送新的request"))
			}
			req, _ := http.NewRequestWithContext(ctx, "POST", url, conn)
			req.Header.Set("ID", id)
			_, err := h2Client.RoundTrip(req)
			if err != nil {
				panic(errors.New("h2Client request发送失败"))
			}
		}()
		// 将从remote收到的写入local
		io.Copy(conn, res.Body)
	}
}
