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

func main() {
	var typeOfApp string
	var configPath string

	flag.StringVar(&typeOfApp, "t", "local", "模式local/remote")
	flag.StringVar(&configPath, "c", "./conf.local.yaml", "配置文件路径")
	flag.Parse()

	switch typeOfApp {
	case "local":
		startLocalService(configPath)
	case "remote":
		startRemoteService(configPath)
	default:
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
