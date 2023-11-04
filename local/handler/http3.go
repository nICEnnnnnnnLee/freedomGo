package handler

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	"github.com/nicennnnnnnlee/freedomGo/utils/geo"
	"github.com/quic-go/quic-go/http3"
)

func HandleHttp3(conn net.Conn, conf *config.Local) {
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
	//remoteAddr := fmt.Sprintf("%s:%d", conf.RemoteHost, conf.RemotePort)
	tlsCfg := &tls.Config{
		InsecureSkipVerify: conf.AllowInsecure,
		ServerName:         conf.HttpDomain,
		VerifyConnection: func(connState tls.ConnectionState) error {
			if conf.AllowInsecure {
				return nil
			}
			return connState.PeerCertificates[0].VerifyHostname(conf.HttpDomain)
		},
	}

	//ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsCfg,
		//QuicConfig: &qconf,
		// StreamHijacker: func(fType http3.FrameType, qConn quic.Connection, qStream quic.Stream, err error) (bool, error) {
		// 	log.Println("fType...", fType)
		// 	return false, err
		// },
	}
	defer roundTripper.Close()
	client := &http.Client{
		Transport: roundTripper,
	}
	// 发送GET请求
	url := "https://" + conf.HttpDomain + conf.HttpPath
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", GenCookie(conf, host, port))
	req.Header.Add("User-Agent", conf.HttpUserAgent)
	//处理返回结果
	res, err := client.Transport.(*http3.RoundTripper).RoundTripOpt(req, http3.RoundTripOpt{DontCloseRequestStream: true})
	// res, err := client.Do(req)
	if err != nil {
		log.Println("client initialize failed...", err.Error())
		panic(err)
		// return
	}
	// log.Println("client 发送完第一次请求...", res.Header.Get("auth"))
	if res.Header.Get("auth") == "ok" {
		// log.Println("client 鉴权成功...", url)
		h, err := res.Body.(http3.HTTPStreamer)
		if !err {
			log.Println("http HTTPStreamer not implmented for http.Response.Body")
			panic(err)
		}
		conn2server := h.HTTPStream()
		if head == "CONNECT" {
			io.WriteString(conn, HttpsProxyEstablished)
		} else {
			io.WriteString(conn2server, header)
		}
		go utils.Pip(conn, conn2server)
		utils.Pip(conn2server, conn)

		// conn2server.Write([]byte("--------------test data from client to server----------------\n"))
		// log.Println("发送测试数据")
		// io.Copy(os.Stdout, conn2server)
	} else {
		log.Println("h3 client 鉴权失败...", url)
		log.Println("res.Status", res.Status)
		log.Println("res.Header", res.Header)
		io.Copy(os.Stdout, res.Body)
		log.Println()
	}
}
