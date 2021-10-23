package handler

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/nicennnnnnnlee/freedomGo/remote/config"
	"github.com/nicennnnnnnlee/freedomGo/utils"
)

var (
	regCookie   = regexp.MustCompile(`Cookie *: *([^\r\n]+)`)
	regDomain   = regexp.MustCompile("my_domain=([^;]+)")
	regPort     = regexp.MustCompile("my_port=([0-9]+)")
	regToken    = regexp.MustCompile("my_token=([^;]+)")
	regUsername = regexp.MustCompile("my_username=([^;]+)")
	regType     = regexp.MustCompile("my_type=([^;]*)")
	regTime     = regexp.MustCompile("my_time=([0-9]+)")
)

func GetAuthorizedConn(authRecv string, conf *config.Remote) net.Conn {

	matches := regCookie.FindStringSubmatch(authRecv)
	if matches == nil {
		panic(utils.ErrHeaderNotRight)
	}
	cookie := matches[1]
	// log.Println("auth cookie received...")
	// log.Println(cookie)
	domain := regDomain.FindStringSubmatch(cookie)[1]
	port := regPort.FindStringSubmatch(cookie)[1]
	token := regToken.FindStringSubmatch(cookie)[1]
	username := regUsername.FindStringSubmatch(cookie)[1]
	typeStr := regType.FindStringSubmatch(cookie)[1]
	timeStr := regTime.FindStringSubmatch(cookie)[1]
	timeInt64, _ := strconv.ParseInt(timeStr, 10, 64)
	if time.Now().UnixMilli()-timeInt64 > 600000 || typeStr != "1" {
		return nil
	}
	// log.Println("auth time valid...")
	pwd, ok := conf.Users[username]
	if ok {
		// log.Println("auth user exists...")
		h := md5.New()
		io.WriteString(h, pwd)
		io.WriteString(h, conf.Salt)
		io.WriteString(h, timeStr)
		exToken := fmt.Sprintf("%x", h.Sum(nil))
		if exToken == token {
			// log.Println("auth user valid...")
			remoteAddr := domain + ":" + port
			conn2server, err := net.Dial("tcp", remoteAddr)
			if err != nil {
				panic(err)
			}

			return conn2server
		}
	}
	return nil
}
