package extend

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Handler for gin
// func HandleTunnel(c *gin.Context) {
// 	Handler(c.Writer, c.Request)
// 	// c.Abort()
// }

// Handler for basic http
func Handler(w http.ResponseWriter, r *http.Request) {
	defer HandleError()
	remoteAddr, err := parseValidAddr(r)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		// fmt.Printf("parseValidAddr err %v\r\n", err.Error())
		return
	}
	ctx := r.Context()
	// conn2server, err := net.Dial("tcp", *remoteAddr)
	conn2server, err := DialTCPContext(ctx, *remoteAddr)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte("connection establish failed: " + *remoteAddr))
		return
	}
	defer conn2server.Close()
	h, result := w.(http.Hijacker)
	if !result {
		w.WriteHeader(400)
		w.Write([]byte("http Hijacker not implmented for http.ResponseWriter"))
		return
	}
	conn2local, brw, err := h.Hijack()
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	if brw.Reader.Buffered() > 0 {
		conn2local.Close()
		return
	}
	conn2local.SetDeadline(time.Time{})
	make101Response(conn2local)
	go Pip(conn2local, conn2server)
	Pip(conn2server, conn2local)
}

func parseValidAddr(r *http.Request) (*string, error) {
	// if r.Proto == "HTTP/2.0" { // HTTP1.1 + HTTP2.0
	// 	return nil, errors.New("proto http2 not allowed on current path")
	// }
	timeStr, err := r.Cookie("my_time")
	if err != nil {
		return nil, errors.New("403 Forbidden")
	}
	typeStr, err := r.Cookie("my_type")
	if err != nil {
		return nil, errors.New("type invalid")
	}
	timeInt64, err := strconv.ParseInt(timeStr.Value, 10, 64)
	if err != nil || time.Now().UnixMilli()-timeInt64 > 600000 || typeStr.Value != "1" {
		return nil, errors.New("timestamp or type invalid")
	}
	username, err := r.Cookie("my_username")
	if err != nil {
		return nil, errors.New("UserName invalid")
	}
	validPassword, ok := Users[username.Value]
	if !ok {
		return nil, errors.New("UserName invalid")
	}
	token, err := r.Cookie("my_token")
	if err != nil {
		return nil, errors.New("token invalid")
	}
	h := md5.New()
	io.WriteString(h, validPassword)
	io.WriteString(h, Salt)
	io.WriteString(h, timeStr.Value)
	exToken := fmt.Sprintf("%x", h.Sum(nil))
	if exToken != token.Value {
		return nil, errors.New("token invalid")
	}
	host, err := r.Cookie("my_domain")
	if err != nil {
		return nil, errors.New("host invalid")
	}
	port, err := r.Cookie("my_port")
	if err != nil {
		return nil, errors.New("port invalid")
	}
	addr := host.Value + ":" + port.Value
	return &addr, nil

}

func make101Response(conn net.Conn) {
	fmt.Fprintf(conn, response_101, RandKey())
	// io.WriteString(conn, response_101)
}

// func make403Response(conn net.Conn) {
// 	io.WriteString(conn, response_403)
// }

func RandKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
