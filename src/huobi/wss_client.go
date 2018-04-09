package main

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

type tickerData struct {
	Open   float64
	Close  float64
	Low    float64
	High   float64
	Amount float64
	Count  int64
	Vol    float64
	Symbol string
}

type marketOverview struct {
	Ch     string
	Ts     int64
	Status string
	Data   []tickerData
}

func huobiWebsocket() {
	u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	log.Println("connecting to ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	c.WriteMessage(websocket.TextMessage, []byte(`{"sub": "market.overview", "id": "id2"}`))

	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			log.Println("read error: ", err)
			continue
		}
		jsonData, err := unzip(data)
		text := string(jsonData)
		if strings.Contains(text, "ping") {
			rspText := strings.Replace(text, "ping", "pong", 1)
			if err = c.WriteMessage(websocket.TextMessage, []byte(rspText)); err != nil {
				log.Println("WriteMessage error: ", err)
				continue
			}
		} else {
			overview := &marketOverview{}
			if err = json.Unmarshal(jsonData, overview); err != nil {
				log.Println("json Unmarshal error: ", err)
				continue
			}
			lock.Lock()
			for _, data := range overview.Data {
				coinClosePrice[data.Symbol] = data
			}
			lock.Unlock()
		}
	}
}
