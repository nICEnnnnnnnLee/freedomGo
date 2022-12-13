package remote

import (
	"fmt"

	"github.com/nicennnnnnnlee/freedomGo/remote/config"
)

func Start(conf *config.Remote) {
	switch conf.HTTPMode {
	case "grpc":
		StartGRPC(conf)
	case "ws":
		StartWs(conf)
	default:
		fmt.Println("HTTPMode 必须为 grpc 或者 ws")
	}
}
