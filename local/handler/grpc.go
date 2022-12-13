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

	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	remoteAddr := fmt.Sprintf("%s:%d", conf.RemoteHost, conf.RemotePort)
	var opts []grpc.DialOption = make([]grpc.DialOption, 0, 1)
	if conf.RemoteSSL {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: conf.AllowInsecure,
			ServerName:         conf.HttpDomain,
			VerifyConnection: func(connState tls.ConnectionState) error {
				if conf.AllowInsecure {
					return nil
				}
				return connState.PeerCertificates[0].VerifyHostname(conf.HttpDomain)
			},
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	if head == "CONNECT" {
		io.WriteString(conn2local, HttpsProxyEstablished)
	} else {
		stream.Send(&pb.FreedomRequest{Data: []byte(header)})
	}

	go func() {
		defer stream.CloseSend()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				// log.Printf("stream.Recv error: %v", err)
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
