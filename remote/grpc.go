package remote

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "github.com/nicennnnnnnlee/freedomGo/grpc"
	"github.com/nicennnnnnnlee/freedomGo/utils"

	"github.com/nicennnnnnnlee/freedomGo/remote/config"
	"github.com/nicennnnnnnlee/freedomGo/remote/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedFreedomServer
}

var conf config.Remote

func (s *server) Pipe(stream pb.Freedom_PipeServer) error {
	defer utils.HandleError()
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.Internal, "missing correct incoming metadata in rpc context")
	}
	cookie, ok := md["cookie"]
	if !ok {
		log.Println(md)
		return status.Errorf(codes.Internal, "missing correct Cookie")
	}
	// log.Println(md)

	remoteAddr := handler.GetRemoteAddrFromCookie(&cookie[0], &conf)
	if remoteAddr == nil {
		return status.Errorf(codes.Internal, "missing correct Cookie")
	}
	conn2server := handler.GetRemoteConn(remoteAddr, &conf)
	defer conn2server.Close()
	// 开一个协程推流
	go func() {
		defer conn2server.Close()
		defer utils.HandleError()
		for {
			in, err := stream.Recv()
			if err == io.EOF || err != nil {
				break
			}
			if _, err = conn2server.Write(in.Data); err != nil {
				break
			}
		}
	}()
	// 拉流
	for {
		buffer := make([]byte, 1024)
		for {
			len, err := conn2server.Read(buffer)
			if err != nil {
				// log.Println("err:", err)
				return nil
			}
			if len > 0 {
				err = stream.Send(&pb.FreedomResponse{Data: buffer[:len]})
				if err != nil {
					return err
				}
			}
		}
	}
}

func StartGRPC(confRemote *config.Remote) {
	conf = *confRemote
	var opts []grpc.ServerOption = make([]grpc.ServerOption, 0, 1)
	if conf.UseSSL {
		tlsCert := &utils.TlsCert{
			CertPath:       conf.CertPath,
			KeyPath:        conf.KeyPath,
			AttempDuration: time.Minute * 5,
		}
		GetCertFunc := tlsCert.GetCertFunc()
		opt := credentials.NewTLS(&tls.Config{GetCertificate: GetCertFunc})
		opts = append(opts, grpc.Creds(opt))
	}

	fmt.Println("服务器开始监听...")
	addr := fmt.Sprintf("%s:%d", conf.BindHost, conf.BindPort)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen.Close()

	s := grpc.NewServer(opts...)           // 创建gRPC服务器
	pb.RegisterFreedomServer(s, &server{}) // 在gRPC服务端注册服务
	// 启动服务
	err = s.Serve(listen)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}
