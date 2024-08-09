package internal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"net/url"
	"syscall"

	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	utls "github.com/refraction-networking/utls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var optsGrpc []grpc.DialOption

func GetGlobalGrpcConfig(conf *config.Local) []grpc.DialOption {
	if optsGrpc != nil {
		return optsGrpc
	}
	var opts []grpc.DialOption
	if conf.RemoteSSL {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: conf.AllowInsecure,
			ServerName:         conf.HttpDomain,
			NextProtos:         conf.TLSClientHelloNextProtos,
			VerifyConnection: func(connState tls.ConnectionState) error {
				if conf.AllowInsecure {
					return nil
				}
				return connState.PeerCertificates[0].VerifyHostname(conf.HttpDomain)
			},
		}
		creds := NewTLS(tlsConfig, conf)
		// creds := credentials.NewTLS(tlsConfig)
		dialer := grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			conn, err := utils.DialTCPContext(ctx, addr, conf.DNSServer)
			if err != nil {
				return nil, err
			}
			// conn = conf.GetUConn(conn)
			return conn, nil
		})
		opts = []grpc.DialOption{dialer, grpc.WithTransportCredentials(creds)}
	} else {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	opts = append(opts, grpc.WithUserAgent(conf.HttpUserAgent))
	optsGrpc = opts
	return optsGrpc
}

func NewTLS(c *tls.Config, lc *config.Local) credentials.TransportCredentials {
	tc := &tlsCreds{
		conf:      lc,
		_tlsCreds: credentials.NewTLS(c),
	}
	return tc
}

type tlsCreds struct {
	conf      *config.Local
	_tlsCreds credentials.TransportCredentials
}

func (c *tlsCreds) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	// 	// use local cfg to avoid clobbering ServerName if using multiple endpoints
	// cfg := CloneTLSConfig(c.cfg)

	// if cfg.ServerName == "" {
	// 	serverName, _, err := net.SplitHostPort(authority)
	// 	if err != nil {
	// 		// If the authority had no host port or if the authority cannot be parsed, use it as-is.
	// 		serverName = authority
	// 	}
	// 	cfg.ServerName = serverName
	// }
	// conn := tls.Client(rawConn, cfg)
	conn := c.conf.GetUConn(rawConn)
	errChannel := make(chan error, 1)
	go func() {
		errChannel <- conn.Handshake()
		close(errChannel)
	}()
	select {
	case err := <-errChannel:
		if err != nil {
			conn.Close()
			return nil, nil, err
		}
	case <-ctx.Done():
		conn.Close()
		return nil, nil, ctx.Err()
	}
	tlsInfo := TLSInfo{
		State: conn.ConnectionState(),
		CommonAuthInfo: credentials.CommonAuthInfo{
			SecurityLevel: credentials.PrivacyAndIntegrity,
		},
	}
	id := SPIFFEIDFromState(conn.ConnectionState())
	if id != nil {
		tlsInfo.SPIFFEID = id
	}
	return WrapSyscallConn(rawConn, conn), tlsInfo, nil
}

// Sern is closed, it MUST close the net.Conn provided.
func (c *tlsCreds) ServerHandshake(n net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return c._tlsCreds.ServerHandshake(n)
}

func (c *tlsCreds) Info() credentials.ProtocolInfo {
	return c._tlsCreds.Info()
}

func (c *tlsCreds) Clone() credentials.TransportCredentials {
	tc := &tlsCreds{
		_tlsCreds: c._tlsCreds.Clone(),
	}
	return tc
}

func (c *tlsCreds) OverrideServerName(s string) error {
	return c._tlsCreds.OverrideServerName(s)
}

type TLSInfo struct {
	State utls.ConnectionState
	credentials.CommonAuthInfo
	// This API is experimental.
	SPIFFEID *url.URL
}

// AuthType returns the type of TLSInfo as a string.
func (t TLSInfo) AuthType() string {
	return "tls"
}

// GetSecurityValue returns security info requested by channelz.
func (t TLSInfo) GetSecurityValue() credentials.ChannelzSecurityValue {
	v := &credentials.TLSChannelzSecurityValue{
		StandardName: cipherSuiteLookup[t.State.CipherSuite],
	}
	// Currently there's no way to get LocalCertificate info from tls package.
	if len(t.State.PeerCertificates) > 0 {
		v.RemoteCertificate = t.State.PeerCertificates[0].Raw
	}
	return v
}

var cipherSuiteLookup = map[uint16]string{
	tls.TLS_RSA_WITH_RC4_128_SHA:                "TLS_RSA_WITH_RC4_128_SHA",
	tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:           "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
	tls.TLS_RSA_WITH_AES_128_CBC_SHA:            "TLS_RSA_WITH_AES_128_CBC_SHA",
	tls.TLS_RSA_WITH_AES_256_CBC_SHA:            "TLS_RSA_WITH_AES_256_CBC_SHA",
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256:         "TLS_RSA_WITH_AES_128_GCM_SHA256",
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384:         "TLS_RSA_WITH_AES_256_GCM_SHA384",
	tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:        "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:          "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:     "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	tls.TLS_FALLBACK_SCSV:                       "TLS_FALLBACK_SCSV",
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256:         "TLS_RSA_WITH_AES_128_CBC_SHA256",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	tls.TLS_AES_128_GCM_SHA256:                  "TLS_AES_128_GCM_SHA256",
	tls.TLS_AES_256_GCM_SHA384:                  "TLS_AES_256_GCM_SHA384",
	tls.TLS_CHACHA20_POLY1305_SHA256:            "TLS_CHACHA20_POLY1305_SHA256",
}

// internel
// func CloneTLSConfig(cfg *tls.Config) *tls.Config {
// 	if cfg == nil {
// 		return &tls.Config{}
// 	}

// 	return cfg.Clone()
// }

type sysConn = syscall.Conn

type syscallConn struct {
	net.Conn
	// sysConn is a type alias of syscall.Conn. It's necessary because the name
	// `Conn` collides with `net.Conn`.
	sysConn
}

func WrapSyscallConn(rawConn, newConn net.Conn) net.Conn {
	sysConn, ok := rawConn.(syscall.Conn)
	if !ok {
		return newConn
	}
	return &syscallConn{
		Conn:    newConn,
		sysConn: sysConn,
	}
}

func SPIFFEIDFromState(state utls.ConnectionState) *url.URL {
	if len(state.PeerCertificates) == 0 || len(state.PeerCertificates[0].URIs) == 0 {
		return nil
	}
	return SPIFFEIDFromCert(state.PeerCertificates[0])
}

func SPIFFEIDFromCert(cert *x509.Certificate) *url.URL {
	if cert == nil || cert.URIs == nil {
		return nil
	}
	var spiffeID *url.URL
	for _, uri := range cert.URIs {
		if uri == nil || uri.Scheme != "spiffe" || uri.Opaque != "" || (uri.User != nil && uri.User.Username() != "") {
			continue
		}
		// From this point, we assume the uri is intended for a SPIFFE ID.
		if len(uri.String()) > 2048 {
			log.Println("invalid SPIFFE ID: total ID length larger than 2048 bytes")
			return nil
		}
		if len(uri.Host) == 0 || len(uri.Path) == 0 {
			log.Println("invalid SPIFFE ID: domain or workload ID is empty")
			return nil
		}
		if len(uri.Host) > 255 {
			log.Println("invalid SPIFFE ID: domain length larger than 255 characters")
			return nil
		}
		// A valid SPIFFE certificate can only have exactly one URI SAN field.
		if len(cert.URIs) > 1 {
			log.Println("invalid SPIFFE ID: multiple URI SANs")
			return nil
		}
		spiffeID = uri
	}
	return spiffeID
}
