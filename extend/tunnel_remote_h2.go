package extend

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

// Handler for gin
// func HandleTunnel(c *gin.Context) {
// 	Handler(c.Writer, c.Request)
// 	// c.Abort()
// }

var idConnTable = NewIdConnMap()

type IdConnMap struct {
	lock sync.RWMutex
	m    map[string]net.Conn
}

func NewIdConnMap() *IdConnMap {
	return &IdConnMap{m: make(map[string]net.Conn)}
}

func (m *IdConnMap) Set(id string, conn net.Conn) {
	m.lock.Lock()
	m.m[id] = conn
	m.lock.Unlock()
}

func (m *IdConnMap) Delete(id string) {
	m.lock.Lock()
	delete(m.m, id)
	m.lock.Unlock()
}

func (m *IdConnMap) Get(id string) (net.Conn, bool) {
	m.lock.RLock()
	conn, ok := m.m[id]
	m.lock.RUnlock()
	return conn, ok
}

// Handler for basic http
func HandlerH2(w http.ResponseWriter, r *http.Request) {
	defer HandleError()
	// id := r.Header.Get("ID") + ":" + r.RemoteAddr
	id := r.Header.Get("ID")
	// id := r.RemoteAddr
	if id == "" {
		w.WriteHeader(403)
		w.Write([]byte("There should be a id header"))
		log.Println("There should be a id header")
		return
	}
	ctx := r.Context()
	// 如果Method是GET，代表新建连接
	if r.Method == "GET" {
		remoteAddr, err := parseValidAddr(r)
		if err != nil {
			w.WriteHeader(403)
			w.Write([]byte(err.Error()))
			log.Printf("parseValidAddr err %v\r\n", err.Error())
			return
		}
		// conn2server, err := net.Dial("tcp", *remoteAddr)
		conn2server, err := DialTCPContext(ctx, *remoteAddr)
		if err != nil {
			w.WriteHeader(403)
			w.Write([]byte("connection establish failed: " + *remoteAddr))
			return
		}
		defer conn2server.Close()
		idConnTable.Set(id, conn2server)
		log.Println("id:", id, "; addr:", r.RemoteAddr, conn2server.RemoteAddr())
		defer idConnTable.Delete(id)
		w.Header().Add("auth", "ok")
		w.WriteHeader(200)
		f, _ := w.(http.Flusher)
		f.Flush()
		log.Println("connection established to: " + *remoteAddr)
		buffer := make([]byte, 1024)
		for {
			len, err := conn2server.Read(buffer)
			if len > 0 {
				// log.Println("从conn2server中读取数据大小：", len)
				_, err = w.Write(buffer[:len])
				if len < 1024 {
					f.Flush()
				}
			}

			if err != nil {
				panic(err)
			}
		}
	} else {
		// 从map中取出链接
		conn2server, ok := idConnTable.Get(id)
		if !ok {
			w.WriteHeader(403)
			w.Write([]byte("No connection found in table"))
			log.Println("No connection found in table", id)
			return
		}
		log.Println("id:", id, "; addr:", r.RemoteAddr, conn2server.RemoteAddr())
		if r.Header.Get("Next") != "1" {
			defer conn2server.Close()
		}
		io.Copy(conn2server, r.Body)
	}
}
