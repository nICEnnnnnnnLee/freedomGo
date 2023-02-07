package local

import (
	"fmt"
	"log"
	"net"

	"github.com/nicennnnnnnlee/freedomGo/extend"
	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/local/handler"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	"github.com/nicennnnnnnlee/freedomGo/utils/geo"
)

func handleClient(conn net.Conn, conf *config.Local) {
	defer conn.Close()
	defer utils.HandleError()
	switch conf.ProxyType {
	case config.HTTP:
		// handler.HandleHttp(conn, conf)
		switch conf.ProxyMode {
		case "grpc":
			handler.HandleGrpc(conn, conf)
		case "ws":
			handler.HandleWebSocket(conn, conf)
		case "http2":
			handler.HandleHttp2(conn, conf)
		default:
			log.Fatalf("ProxyMode 只能是ws http2 grpc, 当前为 %s\n", conf.ProxyMode)
		}

	case config.SOCKS5:
		switch conf.ProxyMode {
		case "grpc":
			handler.HandleSocks5_GRPC(conn, conf)
		case "ws":
			handler.HandleSocks5(conn, conf)
		// case "http2":
		// 	handler.HandleHttp2(conn, conf)
		default:
			log.Fatalf("ProxyMode 只能是ws grpc, 当前为 %s\n", conf.ProxyMode)
		}
	default:
		log.Fatalf("ProxyType 只能是%s 或 %s, 当前为 %s\n", config.HTTP, config.SOCKS5, conf.ProxyType)
	}
	//
}

func Start(conf *config.Local) {
	initGeoConfig(conf)
	fmt.Println("服务器开始监听...")
	addr := fmt.Sprintf("%s:%d", conf.BindHost, conf.BindPort)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen.Close()
	// 初始化extend配置(对于http2来说有用)
	extend.DnsResolver = conf.DNSServer
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Accept err:", err)
		} else {
			go handleClient(conn, conf)
		}
	}
}

func initGeoConfig(conf *config.Local) {
	geoConf := conf.GeoDomain
	if geoConf == nil {
		return
	}
	geo.InitProxySet(geoConf.GfwPath)
	geo.InitDirectSet(geoConf.DirectPath)
}
