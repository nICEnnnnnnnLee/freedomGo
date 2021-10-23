package geo

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"strings"
)

type void struct{}

var (
	empty             void
	proxySet          = make(map[string]void)
	directSet         = make(map[string]void)
	ptrTrue, ptrFalse = new(bool), new(bool)
)

func init() {
	*ptrTrue = true
	*ptrFalse = false
}
func InitProxySet(gfwListPath string) {

	content, err := ioutil.ReadFile(gfwListPath)

	if err != nil {
		log.Println("gfwListPath is not right")
		return
	}
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(content)))
	_, err = base64.StdEncoding.Decode(dbuf, content)
	if err != nil {
		log.Println("gfwList format not right")
		return
	}
	lines := strings.Split(string(dbuf), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Compare(line, "!##############General List End#################") == 0 {
			break
		}
		if strings.Contains(line, ".*") {
			continue
		} else if strings.Contains(line, "*") {
			line = strings.ReplaceAll(line, "*", "/")
		}

		if strings.HasPrefix(line, "||") {
			line = line[2:]
		} else if strings.HasPrefix(line, "|") || strings.HasPrefix(line, ".") {
			line = line[1:]
		}

		if strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "@") {
			continue
		}

		start := strings.Index(line, "://")
		if start > -1 {
			line = line[start+3:]
		}
		end := strings.Index(line, "/")
		if end > -1 {
			line = line[:end]
		}
		proxySet[line] = empty
	}
	// log.Println("ProxySet size:", len(proxySet))

}

func InitDirectSet(cnListPath string) {

	content, err := ioutil.ReadFile(cnListPath)

	if err != nil {
		log.Println("cnListPath is not right")
		return
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		directSet[line] = empty
	}
	// log.Println("DirectSet size:", len(directSet))
}

func IsDirect(host string) *bool {

	if strings.HasSuffix(host, ".cn") {
		return ptrTrue
	}
	idx1 := strings.LastIndex(host, ".")
	idx2 := strings.LastIndex(host[:idx1], ".")
	for {
		suffix := host[idx2+1:]
		// log.Println("suffix: ", suffix)
		if _, ok := directSet[string(suffix)]; ok {
			return ptrTrue
		} else if _, ok := proxySet[string(suffix)]; ok {
			return ptrFalse
		} else if idx2 == -1 {
			return nil
		}
		idx2 = strings.LastIndex(host[:idx2], ".")
	}
}
