package config

import (
	"bytes"
	"fmt"
	"reflect"
)

const (
	SOCKS5    = "socks5"
	HTTP      = "http"
	WEBSOCKET = "ws"
)

type Local struct {
	ProxyType                string     `yaml:"ProxyType"`
	ProxyMode                string     `yaml:"ProxyMode"`
	BindHost                 string     `yaml:"BindHost"`
	BindPort                 uint16     `yaml:"BindPort"`
	DNSServer                string     `yaml:"DNSServer"`
	RemoteHost               string     `yaml:"RemoteHost"`
	RemotePort               uint16     `yaml:"RemotePort"`
	RemoteSSL                bool       `yaml:"RemoteSSL"`
	TLSClientHelloID         string     `yaml:"ClientHelloID"`
	TLSClientHelloRawPath    string     `yaml:"ClientHelloRawPath"`
	TLSClientHelloNextProtos []string   `yaml:"ClientHelloNextProtos"`
	GeoDomain                *GeoDomain `yaml:"GeoDomain"`
	Salt                     string     `yaml:"Salt"`
	Username                 string     `yaml:"Username"`
	Password                 string     `yaml:"Password"`
	AllowInsecure            bool       `yaml:"AllowInsecure"`
	AllowCertTimeOutdated    bool       `yaml:"AllowCertTimeOutdated"`

	HttpPath      string `yaml:"HttpPath"`
	HttpDomain    string `yaml:"HttpDomain"`
	HttpUserAgent string `yaml:"HttpUserAgent"`

	GrpcServiceName string `yaml:"GrpcServiceName"`
}

func (s *Local) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s.ProxyMode = WEBSOCKET
	s.GrpcServiceName = "freedomGo.grpc.Freedom"
	s.AllowCertTimeOutdated = false
	s.TLSClientHelloID = "go"
	s.TLSClientHelloNextProtos = []string{"http/1.1"}
	type plain Local
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	s.InitTlsCfg()
	s.InitTlsSpec()
	return nil
}

type GeoDomain struct {
	DirectIfNotInRules bool   `yaml:"DirectIfNotInRules"`
	GfwPath            string `yaml:"GfwPath"`
	DirectPath         string `yaml:"DirectPath"`
}

func (local *Local) String() string {
	// 如果为空，直接返回
	if local == nil {
		return "<nil>"
	}
	typ := reflect.TypeOf(local).Elem()
	obj := reflect.ValueOf(local).Elem()
	numField := typ.NumField()
	buffer := bytes.NewBufferString("\n")
	for i := 0; i < numField; i++ {
		key := typ.Field(i).Name
		value := obj.Field(i)
		fmt.Fprintf(buffer, "%v:\t%v\n", key, value)
		// fmt.Println(key, value)
	}
	return buffer.String()
}

func (geoDomain *GeoDomain) String() string {
	// 如果为空，直接返回
	if geoDomain == nil {
		return "<nil>"
	}
	typ := reflect.TypeOf(geoDomain).Elem()
	obj := reflect.ValueOf(geoDomain).Elem()
	numField := typ.NumField()
	buffer := bytes.NewBufferString("\n")
	for i := 0; i < numField; i++ {
		key := typ.Field(i).Name
		value := obj.Field(i)
		fmt.Fprintf(buffer, "\t%v:\t%v\n", key, value)
		// fmt.Println(key, value)
	}
	return buffer.String()
}
