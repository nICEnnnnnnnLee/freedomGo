package extend

import (
	"fmt"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

// Handler for gin
// func HandleTunnel(c *gin.Context) {
// 	Handler(c.Writer, c.Request)
// 	// c.Abort()
// }

// Handler for basic http
func HandlerH3(w http.ResponseWriter, r *http.Request) {
	defer HandleError()
	remoteAddr, err := parseValidAddr(r)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		fmt.Printf("parseValidAddr err %v\r\n", err.Error())
		return
	}
	ctx := r.Context()
	// conn2server, err := net.Dial("tcp", *remoteAddr)
	conn2server, err := DialTCPContext(ctx, *remoteAddr)
	if err != nil {
		w.WriteHeader(502)
		w.Write([]byte("connection establish failed: " + *remoteAddr))
		return
	}
	defer conn2server.Close()
	h, result := r.Body.(http3.HTTPStreamer)
	if !result {
		w.WriteHeader(400)
		w.Write([]byte("http HTTPStreamer not implmented for http.Request.Body"))
		fmt.Println("http HTTPStreamer not implmented for http.Request.Body")
		return
	}
	w.Header().Set("auth", "ok")
	w.WriteHeader(200)
	w.(http.Flusher).Flush()
	conn2local := h.HTTPStream()
	// conn2local.SetDeadline(time.Time{})
	// make101Response(conn2local)

	go Pip(conn2local, conn2server)
	Pip(conn2server, conn2local)
}
