package handler

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"

	"github.com/nicennnnnnnlee/freedomGo/local/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
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
