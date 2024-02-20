package extend

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

// Handler for gin
// func HandleTunnel(c *gin.Context) {
// 	Handler(c.Writer, c.Request)
// 	// c.Abort()
// }

var upgrader = websocket.Upgrader{} // use default options
// Handler for basic http
func HandlerRealWs(w http.ResponseWriter, r *http.Request) {
	defer HandleError()
	remoteAddr, err := parseValidAddr(r)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	ctx := r.Context()
	// conn2server, err := net.Dial("tcp", *remoteAddr)
	conn2server, err := DialTCPContext(ctx, *remoteAddr)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte("connection establish failed: " + *remoteAddr))
		return
	}
	defer conn2server.Close()

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte("fail to upgrade to ws: " + *remoteAddr))
		return
	}
	go client2Remote(c, conn2server)
	remote2Client(conn2server, c)
}

func remote2Client(r net.Conn, c *websocket.Conn) {
	defer c.Close()
	defer r.Close()
	defer HandleError()
	buffer := make([]byte, 1024)
	for {
		len, err := r.Read(buffer)
		if len > 0 {
			err = c.WriteMessage(websocket.BinaryMessage, buffer[:len])
			if err != nil {
				// panic(err)
				break
			}
		}
		if err != nil {
			// panic(err)
			break
		}
	}
}
func client2Remote(c *websocket.Conn, r net.Conn) {
	defer c.Close()
	defer r.Close()
	defer HandleError()
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			// panic(err)
			break
		}
		if mt == websocket.BinaryMessage {
			_, err = r.Write(msg)
			if err != nil {
				// panic(err)
				break
			}
		}
	}
}
