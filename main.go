package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nicennnnnnnlee/freedomGo/grpc"
	"github.com/nicennnnnnnlee/freedomGo/local"
	lconf "github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/remote"
	rconf "github.com/nicennnnnnnlee/freedomGo/remote/config"

	"gopkg.in/yaml.v2"
)

var (
	version   = "Unknown"
	buildTime = "Unknown"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("FreedomGo")
	fmt.Println("\tVersion:\t", version)
	fmt.Println("\tBuildTime:\t", buildTime)

	var typeOfApp string
	var configPath string
	flag.StringVar(&typeOfApp, "type", "local", "模式local/remote")
	flag.StringVar(&typeOfApp, "t", "local", "模式local/remote")
	flag.StringVar(&configPath, "config", "./conf.local.yaml", "配置文件路径")
	flag.StringVar(&configPath, "c", "./conf.local.yaml", "配置文件路径")
	flag.Parse()
	// 尝试读取环境变量的配置
	val, exist := os.LookupEnv("APP_CONFIG_F0")
	var configByte []byte
	if exist {
		configByte, _ = base64.StdEncoding.DecodeString(val)
	}

	switch typeOfApp {
	case "local":
		startLocalService(configPath, configByte)
	case "remote":
		startRemoteService(configPath, configByte)
	default:
		flag.Usage()
		fmt.Println("仅支持local或remote模式")
	}

}

func startLocalService(path string, configByte []byte) {
	var conf lconf.Local
	// out, _ := yaml.Marshal(config.New())
	// fmt.Println(string(out))

	if configByte != nil {
		err := yaml.Unmarshal(configByte, &conf)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		yamlData, err := os.ReadFile(path)
		if err != nil {
			log.Fatalln(err)
		}
		err = yaml.Unmarshal(yamlData, &conf)
		if err != nil {
			log.Fatalln(err)
		}
	}
	fmt.Printf("%+v\n", &conf)

	grpc.Freedom_ServiceDesc.ServiceName = conf.GrpcServiceName
	grpc.Freedom_Method = "/" + conf.GrpcServiceName + "/Pipe"
	local.Start(&conf)
}

func startRemoteService(path string, configByte []byte) {
	var conf rconf.Remote
	// out, _ := yaml.Marshal(config.NewRemote())
	// fmt.Println(string(out))

	if configByte != nil {
		err := yaml.Unmarshal(configByte, &conf)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		yamlData, err := os.ReadFile(path)
		if err != nil {
			log.Fatalln(err)
		}
		err = yaml.Unmarshal(yamlData, &conf)
		if err != nil {
			log.Fatalln(err)
		}
	}
	fmt.Println(&conf)
	grpc.Freedom_ServiceDesc.ServiceName = conf.GrpcServiceName
	grpc.Freedom_Method = "/" + conf.GrpcServiceName + "/Pipe"
	remote.Start(&conf)
}
