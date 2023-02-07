package extend

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

// // Handler for gin
//
//	func HandleNoroute(c *gin.Context) {
//		if c.Request.Method == "CONNECT" {
//			HandleConnect(c.Writer, c.Request)
//			// c.Abort()
//		}
//	}

// Handler for basic http
func HandleConnectH2(w http.ResponseWriter, r *http.Request) {
	defer HandleError()
	log.Printf("Host: %+v", r.Host)
	// 获取host ip
	dst := strings.Split(r.Host, ":")
	if len(dst) != 2 {
		log.Printf("连接建立失败: %v", dst)
		return
	}
	// 连接远程服务器
	remoteAddr := RemoteHost + ":" + RemotePort
	tlsCfg := &tls.Config{
		InsecureSkipVerify: AllowInsecure,
		ServerName:         HttpDomain,
		NextProtos:         []string{"h2"}, // "http/1.1",
		VerifyConnection: func(connState tls.ConnectionState) error {
			if AllowInsecure {
				return nil
			}
			return connState.PeerCertificates[0].VerifyHostname(HttpDomain)
		},
	}
	ctx := r.Context()
	// conn2server, err := DialTCP(remoteAddr, dnsResolver)
	conn2server, err := DialTLSContext(ctx, "tcp", remoteAddr, tlsCfg)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte("connection establish failed: " + remoteAddr))
		return
	}
	defer conn2server.Close()
	// tlsConn := tls.Client(conn2server, tlsCfg)

	// 劫持获取原始链接，并回复消息
	h, result := w.(http.Hijacker)
	if !result {
		w.WriteHeader(400)
		log.Println("http Hijacker not implmented for http.ResponseWriter")
		return
	}
	conn2local, _, _ := h.Hijack()
	conn2local.SetDeadline(time.Time{})
	defer conn2local.Close()
	io.WriteString(conn2local, HttpsProxyEstablished)

	// 获取h2Client
	transport := &http2.Transport{
		TLSClientConfig: tlsCfg,
		DialTLSContext:  DialTLSContext,
	}
	h2Client, err := transport.NewClientConn(conn2server)
	defer h2Client.Shutdown(ctx)
	if err != nil {
		log.Println("h2Client initialize failed...")
		return
	}
	// 向服务器发送消息，建立tunnel
	req, _ := http.NewRequest("GET", "https://"+HttpDomain+HttpPath, nil)
	req.Header.Set("Cookie", genCookie(dst[0], dst[1]))
	id := fmt.Sprintf("%s:%d", r.Host, time.Now().UnixMilli())
	req.Header.Set("ID", id)
	res, err := h2Client.RoundTrip(req)
	if err != nil {
		log.Println("h2Client initialize failed...")
		return
	}
	// log.Println("h2Client 发送完第一次请求...")
	if res.Header.Get("auth") == "ok" {
		// 将从local收到的发送给 remote
		go func() {
			defer HandleError()
			defer conn2local.Close()
			defer h2Client.Shutdown(ctx)
			if !h2Client.CanTakeNewRequest() {
				panic(errors.New("h2Client 无法再发送新的request"))
			}
			req, _ := http.NewRequest("POST", "https://"+HttpDomain+HttpPath, conn2local)
			req.Header.Set("ID", id)
			_, err := h2Client.RoundTrip(req)
			if err != nil {
				panic(errors.New("h2Client request发送失败"))
			}
		}()
		// 将从remote收到的写入local
		buffer := make([]byte, 1024)
		for {
			len, err := res.Body.Read(buffer)
			// log.Println("从res.Body中读取数据大小：", len)
			if err == nil && len > 0 {
				_, err = conn2local.Write(buffer[:len])
			}
			if err != nil {
				// panic(err)
				break
			}
		}
	}
}

// func HandleHTTPProxyH2(w http.ResponseWriter, r *http.Request) {
// 	firstData := fmt.Sprintf("%s %s %s\r\n", r.Method, r.URL.Path, r.Proto)
// 	for key, values := range r.Header {
// 		for _, value := range values {
// 			firstData += fmt.Sprintf("%s: %s\r\n", key, value)
// 		}
// 	}
// 	firstData += "\r\n"
// 	log.Println(firstData)
// 	defer HandleError()
// 	log.Printf("HTTP Host: %+v", r.Host)
// 	// 获取host ip
// 	dst := strings.Split(r.Host, ":")
// 	var host, port string
// 	host = dst[0]
// 	if len(dst) == 2 {
// 		port = dst[1]
// 	} else {
// 		port = "80"
// 	}
// 	// 连接远程服务器
// 	remoteAddr := RemoteHost + ":" + RemotePort
// 	tlsCfg := &tls.Config{
// 		InsecureSkipVerify: AllowInsecure,
// 		ServerName:         HttpDomain,
// 		NextProtos:         []string{"h2"}, // "http/1.1",
// 		VerifyConnection: func(connState tls.ConnectionState) error {
// 			if AllowInsecure {
// 				return nil
// 			}
// 			return connState.PeerCertificates[0].VerifyHostname(HttpDomain)
// 		},
// 	}
// 	ctx := r.Context()
// 	// log.Printf("连接 remoteAddr: %s, addr: %s\n", remoteAddr, r.Host)
// 	// conn2server, err := DialTCP(remoteAddr, dnsResolver)
// 	conn2server, err := DialTLSContext(ctx, "tcp", remoteAddr, tlsCfg)
// 	if err != nil {
// 		w.WriteHeader(403)
// 		w.Write([]byte("connection establish failed: " + remoteAddr))
// 		return
// 	}
// 	defer conn2server.Close()
// 	// tlsConn := tls.Client(conn2server, tlsCfg)

// 	// 劫持获取原始链接
// 	h, result := w.(http.Hijacker)
// 	if !result {
// 		w.WriteHeader(400)
// 		log.Println("http Hijacker not implmented for http.ResponseWriter")
// 		return
// 	}
// 	conn2local, _, _ := h.Hijack()
// 	conn2local.SetDeadline(time.Time{})
// 	defer conn2local.Close()
// 	// io.WriteString(conn2local, HttpsProxyEstablished)

// 	// 获取h2Client
// 	transport := &http2.Transport{
// 		TLSClientConfig: tlsCfg,
// 		DialTLSContext:  DialTLSContext,
// 	}
// 	h2Client, err := transport.NewClientConn(conn2server)
// 	if err != nil {
// 		log.Println("h2Client initialize failed...")
// 		return
// 	}
// 	// 向服务器发送消息，建立tunnel
// 	req, _ := http.NewRequest("GET", "https://"+HttpDomain+HttpPath, nil)
// 	req.Header.Set("Cookie", genCookie(host, port))
// 	id := fmt.Sprintf("%s:%d", r.Host, time.Now().UnixMilli())
// 	req.Header.Set("ID", id)
// 	res, err := h2Client.RoundTrip(req)
// 	if err != nil {
// 		log.Println("h2Client initialize failed...")
// 		return
// 	}
// 	log.Println("h2Client 发送完第一次请求...")
// 	if res.Header.Get("auth") == "ok" {
// 		// 将从local收到的发送给 remote
// 		go func() {
// 			defer HandleError()
// 			defer conn2local.Close()
// 			if !h2Client.CanTakeNewRequest() {
// 				panic(errors.New("h2Client 无法再发送新的request"))
// 			}

// 			r, w := io.Pipe()
// 			w.Write([]byte(firstData))
// 			go io.Copy(w, conn2local)
// 			if !h2Client.CanTakeNewRequest() {
// 				panic(errors.New("h2Client 无法再发送新的request"))
// 			}
// 			req, _ = http.NewRequest("POST", "https://"+HttpDomain+HttpPath, r)
// 			req.Header.Set("ID", id)
// 			_, err = h2Client.RoundTrip(req)
// 			if err != nil {
// 				panic(errors.New("h2Client request发送失败"))
// 			}
// 			// buffer := make([]byte, 1024)
// 			// for {
// 			// 	len, err := conn2local.Read(buffer)
// 			// 	log.Println("从conn2local中读取数据大小：", len)
// 			// 	if len > 0 {
// 			// 		if !h2Client.CanTakeNewRequest() {
// 			// 			panic(errors.New("h2Client 无法再发送新的request"))
// 			// 		}
// 			// 		req, _ := http.NewRequest("POST", "https://"+HttpDomain+HttpPath, bytes.NewReader(buffer[:len]))
// 			// 		_, err := h2Client.RoundTrip(req)
// 			// 		if err != nil {
// 			// 			panic(errors.New("h2Client request发送失败"))
// 			// 		}
// 			// 	}
// 			// 	if err != nil {
// 			// 		panic(errors.New("从conn2local读取数据失败"))
// 			// 	}
// 			// }
// 		}()
// 		// 将从remote收到的写入local
// 		buffer := make([]byte, 1024)
// 		// len, err := res.Body.Read(buffer)
// 		// log.Println("从res.Body中读取数据大小：", len)
// 		// if len > 1 {
// 		// 	conn2local.Write(buffer[1:len])
// 		// }
// 		// if err != nil {
// 		// 	panic(errors.New("h2Client 接收失败"))
// 		// }
// 		for {
// 			len, err := res.Body.Read(buffer)
// 			// log.Println("从res.Body中读取数据大小：", len)
// 			if len > 0 {
// 				_, err = conn2local.Write(buffer[:len])
// 			}
// 			if err != nil {
// 				// panic(err)
// 				break
// 			}
// 		}
// 	}
// }
