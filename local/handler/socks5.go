package handler

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	pb "github.com/nicennnnnnnlee/freedomGo/grpc"
	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	TypeIpv4   byte = 1
	TypeIpv6   byte = 4
	TypeDomain byte = 3
)

var (
	s5Head           = []byte{0x05, 0x01, 0x00}
	s5HeadRes        = []byte{0x05, 0x00}
	s5ConnRes        = []byte{0x05, 0x00, 0x00, 0x01}
	ErrSock5NotValid = errors.New("not a valid socks5 connection")
)

func HandleSocks5(conn net.Conn, conf *config.Local) {
	// log.Println("receive a socks5 conn...")
	conn2server := genPureConn(conn, conf)
	go utils.Pip(conn, conn2server)
	utils.Pip(conn2server, conn)
}

func genPureConn(conn net.Conn, conf *config.Local) net.Conn {
	var (
		host, port string
	)
	buffer := make([]byte, 128)
	len, err := conn.Read(buffer)
	checkValid(len, 3, err)
	// log.Println("receive socks5 head...")
	if !bytes.Equal(buffer[:3], s5Head) {
		panic(ErrSock5NotValid)
	}
	conn.Write(s5HeadRes)
	len, err = conn.Read(buffer)
	checkValid(len, 0, err)

	switch buffer[1] {
	case TypeIpv4:
		// log.Println("receive socks5 ipv4 info...", len, buffer[:len])
		// host = fmt.Sprintf("%d.%d.%d.%d", buffer[4], buffer[5], buffer[6], buffer[7])
		host = net.IP(buffer[4:8]).String()
		// uport := uint64(buffer[8])<<8 + uint64(buffer[9])
		uport := binary.BigEndian.Uint16(buffer[8:10])
		port = strconv.FormatUint(uint64(uport), 10)
		// fmt.Println("dst addr:", host, port)
	case TypeDomain:
		// log.Println("receive socks5 domain info...", len, buffer[:len])
		domainLen := buffer[4]
		host = string(buffer[5 : 5+domainLen])
		uport := binary.BigEndian.Uint16(buffer[5+domainLen : 7+domainLen])
		port = strconv.FormatUint(uint64(uport), 10)
		// fmt.Println("dst addr:", host, port)
	default:
		panic(ErrSock5NotValid)
	}
	conn2server := GetAuthorizedConn(host, port, conf)

	lhost, lport, _ := net.SplitHostPort(conn2server.LocalAddr().String())
	// rhost, rport, _ := net.SplitHostPort(conn2server.RemoteAddr().String())
	uint64port, _ := strconv.ParseUint(lport, 10, 64)
	// uint16port := uint16(uint64port)
	res := append(s5ConnRes, net.ParseIP(lhost)[12:16]...)
	res = append(res, byte(uint64port>>8), byte(uint64port))
	// fmt.Println("res", res)
	conn.Write(res)
	return conn2server
}

func checkValid(leng int, expectedLen int, err interface{}) {
	if err != nil {
		panic(err)
	}
	if expectedLen > 0 && leng != expectedLen {
		panic(ErrSock5NotValid)
	}
}

func HandleSocks5_GRPC(conn net.Conn, conf *config.Local) {
	var (
		host, port string
	)
	buffer := make([]byte, 128)
	len, err := conn.Read(buffer)
	checkValid(len, 3, err)
	// log.Println("receive socks5 head...")
	if !bytes.Equal(buffer[:3], s5Head) {
		panic(ErrSock5NotValid)
	}
	conn.Write(s5HeadRes)
	len, err = conn.Read(buffer)
	checkValid(len, 0, err)

	switch buffer[1] {
	case TypeIpv4:
		// log.Println("receive socks5 ipv4 info...", len, buffer[:len])
		// host = fmt.Sprintf("%d.%d.%d.%d", buffer[4], buffer[5], buffer[6], buffer[7])
		host = net.IP(buffer[4:8]).String()
		// uport := uint64(buffer[8])<<8 + uint64(buffer[9])
		uport := binary.BigEndian.Uint16(buffer[8:10])
		port = strconv.FormatUint(uint64(uport), 10)
		// fmt.Println("dst addr:", host, port)
	case TypeDomain:
		// log.Println("receive socks5 domain info...", len, buffer[:len])
		domainLen := buffer[4]
		host = string(buffer[5 : 5+domainLen])
		uport := binary.BigEndian.Uint16(buffer[5+domainLen : 7+domainLen])
		port = strconv.FormatUint(uint64(uport), 10)
		// fmt.Println("dst addr:", host, port)
	default:
		panic(ErrSock5NotValid)
	}

	// 建立与remote服务器的连接，得到stream
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
	defer conn.Close()

	lhost, uint64port, _ := conf.BindHost, uint16(conf.BindPort), 0
	// lhost, lport, _ := net.SplitHostPort(client.LocalAddr().String())
	// rhost, rport, _ := net.SplitHostPort(conn2server.RemoteAddr().String())
	// uint64port, _ := strconv.ParseUint(lport, 10, 64)
	// uint16port := uint16(uint64port)
	res := append(s5ConnRes, net.ParseIP(lhost)[12:16]...)
	res = append(res, byte(uint64port>>8), byte(uint64port))
	// fmt.Println("res", res)
	conn.Write(res)

	// 接下来就是充当管道工了
	go func() {
		defer stream.CloseSend()
		defer conn.Close()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {

				break
			}
			conn.Write(req.GetData())
		}
	}()

	buffer = make([]byte, 1024)
	for {
		len, err := conn.Read(buffer)
		if len > 0 {
			stream.Send(&pb.FreedomRequest{Data: buffer[:len]})
		}
		if err != nil {
			panic(err)
		}
	}
}
