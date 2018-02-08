package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gosuri/uilive"
)

var addr = flag.String("addr", "api.huobi.pro", "http service address")
var tgToken = flag.String("tgToken", "", "telegram token")
var second = flag.Int("second", 1, "telegram send drution")

var coinClosePrice = make(map[string]tickerData)
var lock sync.RWMutex

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

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	log.Println("connecting to ", u.String())

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
				log.Println("read error: ", err)
				return
			}
			jsonData, err := unzip(data)
			text := string(jsonData)
			if strings.Contains(text, "ping") {
				rspText := strings.Replace(text, "ping", "pong", 1)
				if err = c.WriteMessage(websocket.TextMessage, []byte(rspText)); err != nil {
					log.Println("WriteMessage error: ", err)
				}
			} else {
				overview := &marketOverview{}
				if err = json.Unmarshal(jsonData, overview); err != nil {
					log.Println("json Unmarshal error: ", err)
				}
				lock.Lock()
				for _, data := range overview.Data {
					coinClosePrice[data.Symbol] = data
				}
				lock.Unlock()
			}
		}
	}()

	ticker := time.NewTicker(time.Duration(*second) * time.Second)
	defer ticker.Stop()

	var writer *uilive.Writer
	if *tgToken == "" {
		writer = uilive.New()
		writer.Start()
		defer writer.Stop()
	}

	for {
		select {
		case _ = <-ticker.C:
			usdtCoins := []string{
				"btc", "bch", "eth", "etc", "ltc",
				"eos", "xrp", "omg", "zec", "neo", "dash",
				"ht", "qtum", "hsr", "dta", "let",
				"snt", "cvc", "smt", "ven", "elf", "xem"}

			lock.RLock()
			text := ""
			for _, coin := range usdtCoins {
				text += getCoinText("usdt", coin, "%.2f")
			}

			lock.RUnlock()

			if writer == nil {
				mdText := "*USDT*\n```\n" + text + "\n```"
				sendTG(mdText)
			} else {
				fmt.Fprintln(writer, text)
			}

		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write error : ", err)
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
