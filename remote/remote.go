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
	case "ws_real":
		StartWsReal(conf)
	default:
		fmt.Println("HTTPMode 必须为 grpc 或者 ws 或者 ws_real 或者 http2 或者 http3")
	}
}
