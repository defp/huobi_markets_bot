package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"strings"
)

var addr = flag.String("addr", "api.huobi.pro", "http service address")

type Ticker struct {
	Open   float64
	Close  float64
	Low    float64
	Hight  float64
	Amount float64
	Count  int64
	Vol    float64
	Symbol string
}

type MarketOverview struct {
	Ch     string
	Ts     int64
	Status string
	Data   []Ticker
}

func unzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)

		c.WriteMessage(websocket.TextMessage, []byte(`{"sub": "market.overview", "id": "id2"}`))

		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			jsonData, err := unzip(data)
			text := string(jsonData)
			if strings.Contains(text, "ping") {
				rspText := strings.Replace(text, "ping", "pong", 1)
				if err = c.WriteMessage(websocket.TextMessage, []byte(rspText)); err != nil {
					log.Println(err)
				}
			} else {
				overview := &MarketOverview{}
				json.Unmarshal(jsonData, overview)
				//log.Println("receive text: ", text)
				log.Println(overview.Ch, overview.Ts)

				for _, data := range overview.Data {
					log.Println(data.Symbol, data.Close)
				}
			}
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			//c.WriteMessage(websocket.TextMessage, []byte(`{"req": "market.xrpusdt.trade.detail", "id": "id1"}`))
			//c.WriteMessage(websocket.TextMessage, []byte(`{"req": "market.xrpusdt.detail", "id": "id2"}`))
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}
