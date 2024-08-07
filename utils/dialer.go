package utils

import (
	"context"
	"net"
)

// 通过指定的dnsserver查询dns, 返回和address的TCP连接
// e.g.	address www.baidu.com   dnsserver 8.8.8.8:53
//
//	Windows下，会直接使用调用系统的dns查询，不会生效
//	其它系统，会从8.8.8.8:53查询到对应ip，并与之建立TCP连接
//
// 以下为解释：
// https://pkg.go.dev/net#hdr-Name_Resolution
// Windows下总是会调用C库，所以无法指定
//
// https://github.com/golang/go/issues/22846
// 容器里需要注意新建 /etc/nsswitch.conf文件
//
//		echo "hosts: files dns" > /etc/nsswitch.conf
//
//		Android Termux, 如果不指定dnsserver, 则默认会去读取/etc/resolv.conf
//	 问题是该文件不存在, 或者是没有权限读取, 默认dns server [::1]:53, 那么肯定失败
//
// echo "nameserver 114.114.114.114" > /etc/resolv.conf
func DialTCP(address string, dnsserver string) (net.Conn, error) {
	var dialer net.Dialer
	if dnsserver != "" {
		dialer.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.DialContext(ctx, "udp", dnsserver)
			},
		}
	}
	return dialer.Dial("tcp", address)
}

func DialTCPContext(ctx context.Context, address string, dnsserver string) (net.Conn, error) {
	var dialer net.Dialer
	if dnsserver != "" {
		dialer.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.DialContext(ctx, "udp", dnsserver)
			},
		}
	}
	return dialer.DialContext(ctx, "tcp", address)
}
