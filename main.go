package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"strconv"
)

var (
	hostname = flag.String("h", "", "The RDP server we will connect to")
	username = flag.String("u", "", "Username for the RDP server")
	password = flag.String("p", "", "Password for the RDP server")
)

func getResolution(ws *websocket.Conn) (width int64, height int64) {
	request := ws.Request();
	dtsize := request.FormValue("dtsize")
	sizeparts := strings.Split(dtsize, "x")

	width, _ = strconv.ParseInt(sizeparts[0], 10, 32)
	height, _ = strconv.ParseInt(sizeparts[1], 10, 32)

	if width < 400 {
		height = 400
	} else if width > 1920 {
		width = 1920
	}

	if height < 300 {
		height = 300
	} else if height > 1080 {
		height = 1080
	}

	return width, height
}

func initSocket(ws *websocket.Conn) {
	sendq := make(chan []byte, 100)
	recvq := make(chan []byte, 5)

	width, height := getResolution(ws)
	fmt.Printf("User requested size %d x %d\n", width, height)

	settings := &rdpConnectionSettings{
		hostname,
		username,
		password,
		int(width),
		int(height),
	}

	go rdpconnect(sendq, recvq, settings)

	for {
		buf := <-sendq
		err := websocket.Message.Send(ws, buf)
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}
}

func main() {
	flag.Parse()

	http.Handle("/", websocket.Handler(initSocket))
	fmt.Printf("http://localhost:%d/\n", 4455)
	err := http.ListenAndServe(":4455", nil)
	if err != nil {
		panic("ListenANdServe: " + err.Error())
	}
}
