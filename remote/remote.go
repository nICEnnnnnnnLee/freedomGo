package remote

import (
	"fmt"

	"github.com/nicennnnnnnlee/freedomGo/remote/config"
)

func Start(conf *config.Remote) {
	switch conf.ProxyMode {
	case "grpc":
		StartGRPC(conf)
	case "ws":
		StartWs(conf)
	case "http2":
		StartHttp2(conf)
	case "http3":
		StartHttp3(conf)
	default:
		fmt.Println("HTTPMode 必须为 grpc 或者 ws")
	}
}
