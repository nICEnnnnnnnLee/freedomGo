package geo

import (
	"log"
	"testing"
)

func TestInit(t *testing.T) {

	log.SetFlags(log.Lshortfile)
	InitProxySet(`../../data/gfwlist.txt`)
	InitDirectSet(`../../data/direct_domains.txt`)
	if r := IsDirect("www.test.com"); r != nil {
		t.Error("www.test.com 应该为空", r)
	}
	if r := IsDirect("www.baidu.com"); r == nil || !*r {
		t.Error("www.baidu.com 应该直连", r)
	}
	if r := IsDirect("baidu.com"); r == nil || !*r {
		t.Error("baidu.com 应该直连", r)
	}
	if r := IsDirect("www.google.com"); r == nil || *r {
		t.Error("www.google.com 应该走代理", r)
	}
}
