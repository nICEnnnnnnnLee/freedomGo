package remote

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/nicennnnnnnlee/freedomGo/extend"
	"github.com/nicennnnnnnlee/freedomGo/remote/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
)

func StartWsReal(conf *config.Remote) {
	fmt.Println("服务器开始监听...")

	// 初始化配置
	extend.Users = conf.Users
	extend.Salt = conf.Salt
	extend.HttpPath = conf.ValidHttpPath
	extend.DnsResolver = conf.DNSServer
	http.HandleFunc(conf.ValidHttpPath, extend.HandlerRealWs)

	addr := fmt.Sprintf("%s:%d", conf.BindHost, conf.BindPort)
	var err error
	if conf.UseSSL {
		fmt.Println("使用Real Websocket over TLS")
		tlsCert := &utils.TlsCert{
			CertPath:       conf.CertPath,
			KeyPath:        conf.KeyPath,
			AttempDuration: time.Minute * 5,
		}
		GetCertFunc := tlsCert.GetCertFunc()
		server := &http.Server{
			Addr:      addr,
			Handler:   nil,
			TLSConfig: &tls.Config{GetCertificate: GetCertFunc},
		}
		err = server.ListenAndServeTLS("", "")
		// err = http.ListenAndServeTLS(addr, conf.CertPath, conf.KeyPath, nil)
	} else {
		fmt.Println("使用Real Websocket over TCP")
		err = http.ListenAndServe(addr, nil)
	}
	panic(err)
}
