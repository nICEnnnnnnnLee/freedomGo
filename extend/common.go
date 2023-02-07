package extend

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	// response_403 = "HTTP/1.1 403 Forbidden\r\n" + "Content-Length: 0\r\n" + "Connection: closed\r\n\r\n"
	response_101 = "HTTP/1.1 101 Switching Protocols\r\n" + "auth: ok\r\n" + "Sec-WebSocket-Accept: %s" +
		"\r\nUpgrade: websocket\r\n" + "Connection: Upgrade\r\n\r\n"
)

var (
	Users       = make(map[string]string)
	DnsResolver = getEnvOr("FCK_DNS_RESOLVER", "")

	UserName      = getEnvOr("FCK_USER_NAME", "USER_NAME")
	Password      = getEnvOr("FCK_PASSWORD", "PASSWORD")
	Salt          = getEnvOr("FCK_SALT", "SALT")
	RemoteHost    = getEnvOr("FCK_REMOTE_HOST", "127.0.0.1")
	RemotePort    = getEnvOr("FCK_REMOTE_PORT", "443")
	HttpDomain    = getEnvOr("FCK_REMOTE_DOMAIN", "test.com")
	HttpPath      = getEnvOr("FCK_REMOTE_HTTP_PATH", "/yyuiopk")
	HttpUserAgent = getEnvOr("FCK_REMOTE_HTTP_UA", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0")
	RemoteSSL     = getEnvOr("FCK_REMOTE_HTTPS", "true") != "false"
	AllowInsecure = getEnvOr("FCK_REMOTE_HTTPS_TRUST_ALL", "false") == "false"
)

type Mux struct {
	ServeMux                 *http.ServeMux
	ConnectMethodHandlerFunc *http.HandlerFunc
}

func NewMux(ConnectMethodHandlerFunc *http.HandlerFunc, ServeMux *http.ServeMux) *Mux {
	return &Mux{
		ServeMux,
		ConnectMethodHandlerFunc,
	}
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" && mux.ConnectMethodHandlerFunc != nil {
		(*mux.ConnectMethodHandlerFunc)(w, r)
	} else {
		mux.ServeMux.ServeHTTP(w, r)
	}
}

func (mux *Mux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.ServeMux.HandleFunc(pattern, handler)
}

func (mux *Mux) Handle(pattern string, handler http.Handler) {
	mux.ServeMux.Handle(pattern, handler)
}

func getEnvOr(key string, defaultVal string) string {
	val, exist := os.LookupEnv(key)
	if exist {
		return val
	} else {
		return defaultVal
	}
}

func HandleError() {
	_ = recover()
	// err := recover()
	// if err != nil {
	// 	log.Println(err)
	// 	log.Println(string(debug.Stack()))
	// }
}

func Pip(from io.ReadCloser, to io.WriteCloser) {
	defer from.Close()
	defer to.Close()
	defer HandleError()
	// io.Copy(to, from)
	buffer := make([]byte, 1024)
	for {
		len, err := from.Read(buffer)
		if len > 0 {
			to.Write(buffer[:len])
		}
		if err != nil {
			panic(err)
		}
	}
}

// func DialTCP(address string, dnsResolver string) (net.Conn, error) {
// 	var dialer net.Dialer
// 	dialer.Timeout = time.Second * 5
// 	if dnsResolver != "" {
// 		dialer.Resolver = &net.Resolver{
// 			PreferGo: true,
// 			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
// 				return dialer.DialContext(ctx, "udp", dnsResolver)
// 			},
// 		}
// 	}
// 	return dialer.Dial("tcp", address)
// }

func DialTCPContext(ctx context.Context, address string) (net.Conn, error) {
	var dialer net.Dialer
	dialer.Timeout = time.Second * 5
	if DnsResolver != "" {
		dialer.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx0 context.Context, network, address string) (net.Conn, error) {
				return dialer.DialContext(ctx0, "udp", DnsResolver)
			},
		}
	}
	return dialer.DialContext(ctx, "tcp", address)
}

func DialTLSContext(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
	var dialer = tls.Dialer{
		Config:    cfg,
		NetDialer: &net.Dialer{},
	}
	dialer.Config = cfg
	dialer.NetDialer.Timeout = time.Second * 5
	if DnsResolver != "" {
		dialer.NetDialer.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.DialContext(ctx, "udp", DnsResolver)
			},
		}
	}
	return dialer.DialContext(ctx, network, addr)
}

// func DialTLS(network, addr string, cfg *tls.Config) (net.Conn, error) {
// 	// tcpConn, err := net.Dial(network, addr)
// 	realAddr := RemoteHost + ":" + RemotePort
// 	log.Printf("连接 realAddr: %s, addr: %s\n", realAddr, addr)
// 	tcpConn, err := DialTCP(realAddr, dnsResolver)
// 	if err != nil {
// 		return nil, err
// 	}
// 	tlsConn := tls.Client(tcpConn, cfg)
// 	return tlsConn, nil
// }

// func DialH2TLS(address string, config *tls.Config, dnsResolver string) (net.Conn, error) {
// 	var dialer net.Dialer
// 	dialer.Timeout = time.Second * 5
// 	if dnsResolver != "" {
// 		dialer.Resolver = &net.Resolver{
// 			PreferGo: true,
// 			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
// 				return dialer.DialContext(ctx, "udp", dnsResolver)
// 			},
// 		}
// 	}
// 	return tls.DialWithDialer(&dialer, "tcp", address, config)
// }

// func DialH2TLSContext(ctx context.Context, address string, cfg *tls.Config, dnsResolver string) (net.Conn, error) {
// 	var dialer tls.Dialer
// 	dialer.Config = cfg
// 	dialer.NetDialer.Timeout = time.Second * 5
// 	if dnsResolver != "" {
// 		dialer.NetDialer.Resolver = &net.Resolver{
// 			PreferGo: true,
// 			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
// 				return dialer.DialContext(ctx, "udp", dnsResolver)
// 			},
// 		}
// 	}
// 	return dialer.DialContext(ctx, "tcp", address)
// }
