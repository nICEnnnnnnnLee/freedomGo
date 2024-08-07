package config

import (
	"log"
	"net"
	"os"

	tls "github.com/refraction-networking/utls"
)

var (
	TlsCfg          *tls.Config
	ClientHelloID   tls.ClientHelloID
	ClientHelloSpec *tls.ClientHelloSpec
	clientHelloRaw  []byte
	fingerprinter   *tls.Fingerprinter
)

func (conf *Local) InitTlsCfg() {
	TlsCfg = &tls.Config{
		InsecureSkipVerify:     conf.AllowInsecure,
		InsecureSkipTimeVerify: conf.AllowCertTimeOutdated,
		ServerName:             conf.HttpDomain,
		NextProtos:             conf.TLSClientHelloNextProtos, // "http/1.1",
		VerifyConnection: func(connState tls.ConnectionState) error {
			if conf.AllowInsecure {
				return nil
			}
			return connState.PeerCertificates[0].VerifyHostname(conf.HttpDomain)
		},
	}
}

func (conf *Local) InitTlsSpec() {
	switch conf.TLSClientHelloID {
	case "custom":
		ClientHelloID = tls.HelloCustom
	case "qq":
		ClientHelloID = tls.HelloQQ_Auto
	case "safari":
		ClientHelloID = tls.HelloSafari_Auto
	case "edge":
		ClientHelloID = tls.HelloEdge_Auto
	case "360":
		ClientHelloID = tls.Hello360_Auto
	case "android_okhttp":
		ClientHelloID = tls.HelloAndroid_11_OkHttp
	case "random":
		ClientHelloID = tls.HelloRandomized
	case "firefox":
		ClientHelloID = tls.HelloFirefox_Auto
	case "ios":
		ClientHelloID = tls.HelloIOS_Auto
	case "chrome":
		ClientHelloID = tls.HelloChrome_Auto
	case "go":
		ClientHelloID = tls.HelloGolang
	default:
		log.Println("no match found for ClientHelloID: ", conf.TLSClientHelloID)
		ClientHelloID = tls.HelloGolang
	}

	if ClientHelloID == tls.HelloCustom {
		rawCapturedClientHelloBytes, err := os.ReadFile(conf.TLSClientHelloRawPath)
		if err != nil {
			panic(err)
		}
		fingerprinter = &tls.Fingerprinter{}
		spec, err := fingerprinter.FingerprintClientHello(rawCapturedClientHelloBytes)
		if err != nil {
			log.Panicf("fingerprinting failed: %v", err)
		}
		ClientHelloSpec = spec
		clientHelloRaw = rawCapturedClientHelloBytes
	} else {
		spec, err := tls.UTLSIdToSpec(ClientHelloID)
		if err == nil {
			ClientHelloSpec = &spec
		}
	}
	if ClientHelloSpec != nil {
		for _, ext := range ClientHelloSpec.Extensions {
			alpnExt, ok := ext.(*tls.ALPNExtension)
			if ok {
				log.Println("原始Alpn: ", alpnExt.AlpnProtocols)
				alpnExt.AlpnProtocols = TlsCfg.NextProtos
				log.Println("修改后Alpn: ", alpnExt.AlpnProtocols)
			}
			asExt, ok := ext.(*tls.ApplicationSettingsExtension)
			if ok {
				log.Println("原始SupportedProtocols: ", asExt.SupportedProtocols)
			}
		}
	}
}

func (conf *Local) GetClientHelloSpec() *tls.ClientHelloSpec {
	var finalSpec *tls.ClientHelloSpec
	if clientHelloRaw != nil {
		spec, _ := fingerprinter.FingerprintClientHello(clientHelloRaw)
		finalSpec = spec
	} else {
		spec, _ := tls.UTLSIdToSpec(ClientHelloID)
		finalSpec = &spec
	}
	for _, ext := range finalSpec.Extensions {
		alpnExt, ok := ext.(*tls.ALPNExtension)
		if ok {
			alpnExt.AlpnProtocols = TlsCfg.NextProtos
		}
		// asExt, ok := ext.(*tls.ApplicationSettingsExtension)
		// if ok {
		// 	asExt.SupportedProtocols = TlsCfg.NextProtos
		// }
	}
	return finalSpec
}

func (conf *Local) GetUConn(tcpConn net.Conn) *tls.UConn {
	var uClient *tls.UConn
	if ClientHelloSpec != nil {
		uClient = tls.UClient(tcpConn, TlsCfg, tls.HelloCustom)
		// uClient.ClientHelloID = ClientHelloID
		spec := conf.GetClientHelloSpec()
		if err := uClient.ApplyPreset(spec); err != nil {
			log.Panicf("applying generated spec failed: %v", err)
		}
	} else {
		uClient = tls.UClient(tcpConn, TlsCfg, ClientHelloID)
	}
	// if err := uClient.Handshake(); err != nil {
	// 	log.Panicf("uConn.Handshake() error: %+v", err)
	// }
	// log.Println("NegotiatedProtocol", uClient.ConnectionState().NegotiatedProtocol)
	return uClient
}
