package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	readWait  = 60 * time.Second
)

var wg sync.WaitGroup

type Response struct {
	ID      int64  `json:"id"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Body64  string `json:"body"`
}

func Requester(ws *websocket.Conn) {
	defer wg.Done()
	d := time.NewTicker(30 * time.Second)
	for {
		select {
		case tm := <-d.C:
			// WRITE REQUESTS HERE
			ws.SetWriteDeadline(time.Now().Add(writeWait))

			nowUnix := time.Now().Unix()
			subscribe := `{"id": `
			subscribe += strconv.FormatInt(nowUnix, 10) // 10 - BASE NUMBER OR DECIMAL, 61-HEXADECIMAL
			subscribe += `,"method": "GET","target": "/v2/market/ticker","body": "eyJwYWlyIjoidGVuX2J0YyJ9"}`
			err := ws.WriteMessage(websocket.TextMessage, []byte(subscribe))
			if err != nil {
				log.Fatal("sending msg:", err)
			}
			defer ws.Close()
			// END WRITE REQ
			fmt.Println("The current time is: ", tm)
		} // END SELECT

	} // END FOR

}

func Reader(ws *websocket.Conn) {
	defer wg.Done()
	for {
		// READ REQUEST
		ws.SetReadDeadline(time.Now().Add(readWait))
		_, m, err := ws.ReadMessage()
		if err != nil {
			log.Fatal("receive msg:", err)
		}

		// UNMARSHALL
		r := &Response{}
		err = json.Unmarshal(m, r)
		if err != nil {
			log.Fatalln("error:", err)
		}

		fmt.Println("ID:", r.ID)
		fmt.Println("Code:", r.Code)

		// DECODE BASE64
		body, _ := b64.StdEncoding.DecodeString(r.Body64)
		fmt.Println("Body:", string(body))

		time.Sleep(5 * time.Second)
	} // END FOR

}

func main() {
	u := url.URL{Scheme: "wss", Host: "api.tokenomy.com", Path: "/v2/ws"}
	log.Printf("connecting to %s", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	wg.Add(2)
	go Requester(ws)
	go Reader(ws)
	wg.Wait()
}
