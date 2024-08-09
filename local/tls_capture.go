package local

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/open-ch/ja3"
	// tls "github.com/refraction-networking/utls"
)

func Capture(bindAddr string) {
	fmt.Println("服务器开始捕捉TLS数据...")
	if strings.HasSuffix(bindAddr, ".yaml") {
		bindAddr = "127.0.0.1"
	}
	addr := fmt.Sprintf("%s:%d", bindAddr, 443)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen.Close()
	cnt := 1
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Accept err:", err)
		} else {
			handleCapture(conn, cnt)
			cnt += 1
			if cnt >= 3 {
				break
			}
		}
	}
}

func handleCapture(conn net.Conn, cnt int) {
	defer conn.Close()
	buffer := make([]byte, 2048)
	len, err := conn.Read(buffer)
	if err != nil {
		return
	}
	cHello := buffer[:len]

	j, err := ja3.ComputeJA3FromSegment(cHello)
	if err != nil {
		return
	}

	ja3Hash := j.GetJA3Hash()
	ja3String := j.GetJA3String()
	sni := j.GetSNI()
	fmt.Printf("JA3Hash: %v\nJA3String: %v\nSNI: %v\n-------------\n", ja3Hash, ja3String, sni)

	// cHelloInfo, err := readClientHello(cHello)
	// if err != nil {
	// 	return
	// }
	// fmt.Printf("ServerName %d: %v\n", cnt, cHelloInfo.ServerName)
	// fmt.Printf("SupportedProtos %d: %v\n", cnt, cHelloInfo.SupportedProtos)
	path := fmt.Sprintf("client_hello_%d_%s.data", cnt, sni)
	os.WriteFile(path, cHello, 0644)
}

// type readOnlyConn struct {
// 	reader io.Reader
// }

// func (conn readOnlyConn) Read(p []byte) (int, error)         { return conn.reader.Read(p) }
// func (conn readOnlyConn) Write(p []byte) (int, error)        { return 0, io.ErrClosedPipe }
// func (conn readOnlyConn) Close() error                       { return nil }
// func (conn readOnlyConn) LocalAddr() net.Addr                { return nil }
// func (conn readOnlyConn) RemoteAddr() net.Addr               { return nil }
// func (conn readOnlyConn) SetDeadline(t time.Time) error      { return nil }
// func (conn readOnlyConn) SetReadDeadline(t time.Time) error  { return nil }
// func (conn readOnlyConn) SetWriteDeadline(t time.Time) error { return nil }

// func readClientHello(cHello []byte) (*tls.ClientHelloInfo, error) {
// 	var hello *tls.ClientHelloInfo
// 	reader := bytes.NewReader(cHello)

// 	err := tls.Server(readOnlyConn{reader: reader}, &tls.Config{
// 		GetConfigForClient: func(argHello *tls.ClientHelloInfo) (*tls.Config, error) {
// 			hello = new(tls.ClientHelloInfo)
// 			*hello = *argHello
// 			return nil, nil
// 		},
// 	}).Handshake()

// 	if hello == nil {
// 		return nil, err
// 	}

// 	return hello, nil
// }
