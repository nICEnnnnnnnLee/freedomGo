package handler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"

	pb "github.com/nicennnnnnnlee/freedomGo/grpc"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	"github.com/nicennnnnnnlee/freedomGo/utils/geo"

	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/local/handler/internal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func HandleGrpc(conn2local net.Conn, conf *config.Local) {
	header, err := utils.ReadHeader(conn2local)
	if err != nil {
		panic(err)
	}
	reg := regexp.MustCompile(`(CONNECT|Host:) ([^ :\r\n]+)(?::(\d+))?`)
	matches := reg.FindStringSubmatch(header)
	if matches == nil {
		panic(utils.ErrHeaderNotRight)
	}
	head, host, port := matches[1], matches[2], matches[3]
	if port == "" {
		port = "80"
	}

	if conf.GeoDomain != nil {
		r := geo.IsDirect(host)
		if (r == nil && conf.GeoDomain.DirectIfNotInRules) ||
			(r != nil && *r) {
			log.Printf("直连 %s: %s\n", host, port)
			handleDirect(conf, host, port, head, conn2local, header)
			return
		}
	}
	handleProxy(conf, host, port, head, conn2local, header)
}

func handleDirect(conf *config.Local, host string, port string, head string, conn2local net.Conn, header string) {
	conn2server := getDirectConn(host, port, conf)
	if head == "CONNECT" {
		io.WriteString(conn2local, HttpsProxyEstablished)
		// conn2local.Write([]byte(HttpsProxyEstablished))
	} else {
		io.WriteString(conn2server, header)
		// conn2server.Write([]byte(header))
	}
	go utils.Pip(conn2local, conn2server)
	utils.Pip(conn2server, conn2local)
}

func handleProxy(conf *config.Local, host string, port string, head string, conn2local net.Conn, header string) {
	remoteAddr := fmt.Sprintf("%s:%d", conf.RemoteHost, conf.RemotePort)
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
		creds := internal.NewTLS(tlsConfig, conf)
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

	conn2server, err := grpc.Dial(remoteAddr, opts...)
	if err != nil {
		log.Panicf("did not connect: %v", err)
	}
	defer conn2server.Close()
	client := pb.NewFreedomClient(conn2server)

	md := metadata.Pairs(
		"Cookie", GenCookie(conf, host, port),
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := client.Pipe(ctx)
	if err != nil {
		panic(err)
	}
	defer stream.CloseSend()
	defer conn2local.Close()
	if head == "CONNECT" {
		io.WriteString(conn2local, HttpsProxyEstablished)
	} else {
		stream.Send(&pb.FreedomRequest{Data: []byte(header)})
	}

	go func() {
		defer stream.CloseSend()
		defer conn2local.Close()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {

				break
			}
			conn2local.Write(req.GetData())
		}
	}()

	buffer := make([]byte, 1024)
	for {
		len, err := conn2local.Read(buffer)
		if len > 0 {
			stream.Send(&pb.FreedomRequest{Data: buffer[:len]})
		}
		if err != nil {
			panic(err)
		}
	}
}
