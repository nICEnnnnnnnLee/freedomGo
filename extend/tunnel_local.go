package extend

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	tls "github.com/refraction-networking/utls"
)

const (
	HttpsProxyEstablished = "HTTP/1.1 200 Connection Established\r\n\r\n"
)

var (
	randomWSKey = RandKey()
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

// // Handler for gin
// func HandleNoroute(c *gin.Context) {
// 	if c.Request.Method == "CONNECT" {
// 		HandleConnect(c.Writer, c.Request)
// 		// c.Abort()
// 	}
// }

// Handler for basic http
func HandleConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("Host: %+v", r.Host)
	defer HandleError()
	// conn2server, err := DialTCP(r.Host, dnsResolver)
	// strings.Split(r.Host, ":")
	// conn2server, err := DialTCP(r.Host, dnsResolver)
	dst := strings.Split(r.Host, ":")
	if len(dst) != 2 {
		log.Printf("连接建立失败: %v", dst)
		return
	}
	ctx := r.Context()
	conn2server, err := GetAuthorizedConn(ctx, dst[0], dst[1])
	if err != nil {
		log.Printf("连接建立失败:%s %v", r.Host, err)
		return
	}
	defer (*conn2server).Close()
	h, result := w.(http.Hijacker)
	if !result {
		w.WriteHeader(400)
		log.Println("http Hijacker not implmented for http.ResponseWriter")
		return
	}
	conn2local, _, _ := h.Hijack()
	conn2local.SetDeadline(time.Time{})
	io.WriteString(conn2local, HttpsProxyEstablished)
	go Pip(conn2local, *conn2server)
	Pip(*conn2server, conn2local)
}

func genCookie(domain string, port string) string {
	format := "my_type=1; my_domain=%s; my_port=%s; my_username=%s; my_time=%s; my_token=%s"
	timeNow := strconv.FormatInt(time.Now().UnixMilli(), 10)
	h := md5.New()
	io.WriteString(h, Password)
	io.WriteString(h, Salt)
	io.WriteString(h, timeNow)
	token := fmt.Sprintf("%x", h.Sum(nil))
	cookies := fmt.Sprintf(format, domain, port, UserName, timeNow, token)
	return cookies
}

func genHeader(domain string, port string) string {
	cookies := genCookie(domain, port)
	var host string
	if RemotePort == "443" {
		host = HttpDomain
	} else {
		host = fmt.Sprintf("%s:%s", HttpDomain, RemotePort)
	}
	headers := fmt.Sprintf(authHeaders, HttpPath, host, HttpUserAgent, randomWSKey, cookies)
	return headers
}

// 和Remote服务器建立连接，并进行鉴权处理
func GetAuthorizedConn(ctx context.Context, host string, port string) (*net.Conn, error) {

	remoteAddr := fmt.Sprintf("%s:%s", RemoteHost, RemotePort)
	conn2server, err := DialTCPContext(ctx, remoteAddr)
	if err != nil {
		return nil, errors.New("connection failed:" + remoteAddr)
	}
	if RemoteSSL {
		conn2server = tls.UClient(conn2server, &tls.Config{
			InsecureSkipVerify: AllowInsecure,
			ServerName:         HttpDomain,
			VerifyConnection: func(connState tls.ConnectionState) error {
				if AllowInsecure {
					return nil
				}
				return connState.PeerCertificates[0].VerifyHostname(HttpDomain)
			},
		}, tls.HelloRandomized)
	}
	authSend := genHeader(host, port)
	io.WriteString(conn2server, authSend)

	authRecv, err := readHeader(conn2server)
	if err == nil {
		err = checkValid(authRecv)
	}
	if err != nil {
		conn2server.Close()
		return nil, err
	}
	// log.Println("auth success")
	return &conn2server, nil
}

func checkValid(receivedHead *string) error {
	if strings.Contains(*receivedHead, "uth: ok") {
		return nil
	}
	var rsp = strings.Trim(*receivedHead, " \r\n")
	var lines = strings.Split(rsp, "\r\n")
	if len(lines) >= 2 {
		log.Println(lines[0], lines[len(lines)-1])
	}
	return errors.New("response do not contains 'uth: ok'")
}
func readHeader(conn net.Conn) (*string, error) {
	buffer := make([]byte, 1024)
	size, err := conn.Read(buffer)
	if err != nil {
		return nil, errors.New("reading remote response failed, see connection type is http or https or https2")
	}
	str := string(buffer[0:size])
	if strings.HasSuffix(str, "\r\n\r\n") && strings.Contains(str, "uth: ok") {
		return &str, nil
	}
	var rsp = strings.Trim(str, " \r\n")
	var lines = strings.Split(rsp, "\r\n")
	if len(lines) >= 2 {
		log.Println(lines[0], lines[len(lines)-1])
	}
	return &str, errors.New("response do not contains 'uth: ok'")
}

// func Init() error {
// 	http.DefaultClient = &http.Client{Transport: &http.Transport{
// 		TLSClientConfig: &tls.Config{
// 			InsecureSkipVerify: AllowInsecure,
// 			ServerName:         HttpDomain,
// 			VerifyConnection: func(connState tls.ConnectionState) error {
// 				if AllowInsecure {
// 					return nil
// 				}
// 				return connState.PeerCertificates[0].VerifyHostname(HttpDomain)
// 			},
// 		},
// 		DisableCompression: true,
// 		ForceAttemptHTTP2:  false,
// 	}}
// 	return nil
// }

// func GetAuthorizedConn2(host string, port string) (*Connection, error) {
// 	var url string
// 	if RemoteSSL {
// 		url = fmt.Sprintf("https://%s:%s%s", RemoteHost, RemotePort, HttpPath)
// 	} else {
// 		url = fmt.Sprintf("http://%s:%s%s", RemoteHost, RemotePort, HttpPath)
// 	}
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, errors.New("url not valid: " + url)
// 	}
// 	req.Header.Add("User-Agent", HttpUserAgent)
// 	req.Header.Add("Cookie", genCookie(host, port))
// 	rsp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, errors.New("connection failed:" + url)
// 	}
// 	if rsp.StatusCode != 101 || rsp.Header.Get("Auth") != "ok" {
// 		log.Printf("rsp.Status %v", rsp.Status)
// 		io.Copy(os.Stdout, rsp.Body)
// 		fmt.Println()
// 		rsp.Body.Close()
// 		return nil, errors.New("response header do not match 'auth: ok'")
// 	}
// 	conn2remote, _ := rsp.Body.(Connection)
// 	return &conn2remote, nil
// }
