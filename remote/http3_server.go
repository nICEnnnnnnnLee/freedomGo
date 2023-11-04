package remote

import (
	"fmt"
	"net/http"

	"github.com/nicennnnnnnlee/freedomGo/extend"
	"github.com/nicennnnnnnlee/freedomGo/remote/config"
	"github.com/quic-go/quic-go/http3"
)

func StartHttp3(conf *config.Remote) {

	//&http3.RoundTripper{StreamHijacker}
	//hclient := &http.Client{}
	fmt.Println("服务器开始监听...")

	// 初始化配置
	extend.Users = conf.Users
	extend.Salt = conf.Salt
	extend.HttpPath = conf.ValidHttpPath
	extend.DnsResolver = conf.DNSServer
	if conf.HTTP3WebDir != "" {
		http.Handle("/", http.FileServer(http.Dir(conf.HTTP3WebDir)))
	}
	http.HandleFunc(conf.ValidHttpPath, extend.HandlerH3)

	addr := fmt.Sprintf("%s:%d", conf.BindHost, conf.BindPort)
	var err error
	if conf.UseSSL {
		err = http3.ListenAndServeQUIC(addr, conf.CertPath, conf.KeyPath, nil)
	} else {
		fmt.Println("必须使用加密传输")
	}
	panic(err)
}
