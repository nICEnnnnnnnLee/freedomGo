package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

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

	switch typeOfApp {
	case "local":
		startLocalService(configPath)
	case "remote":
		startRemoteService(configPath)
	default:
		flag.Usage()
		fmt.Println("仅支持local或remote模式")
	}

}

func startLocalService(path string) {
	var conf lconf.Local
	// out, _ := yaml.Marshal(config.New())
	// fmt.Println(string(out))

	yamlData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(yamlData, &conf)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", conf)

	local.Start(&conf)
}

func startRemoteService(path string) {
	var conf rconf.Remote
	// out, _ := yaml.Marshal(config.NewRemote())
	// fmt.Println(string(out))

	yamlData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(yamlData, &conf)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(conf)

	remote.Start(&conf)
}
