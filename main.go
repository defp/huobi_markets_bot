package main

import (
	"flag"
	"net/url"
	"os"
	"os/signal"
	"time"

	"encoding/json"
	"strings"

	"github.com/gorilla/websocket"

	"sync"

	"github.com/evalphobia/logrus_sentry"
	log "github.com/sirupsen/logrus"
)

var addr = flag.String("addr", "api.huobi.pro", "http service address")
var tgToken = flag.String("tgToken", "", "telegram token")
var dsn = flag.String("dsn", "", "sentry dsn")
var second = flag.Int("second", 1, "telegram send drution")
var tg = flag.Bool("tg", false, "sendTG mode")

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
	log.SetOutput(os.Stdout)

	flag.VisitAll(func(i *flag.Flag) {
		log.Info(i.Name, "  ", i.Value)
	})

	if (*dsn) != "" {
		hook, err := logrus_sentry.NewSentryHook(*dsn, []log.Level{
			log.PanicLevel, log.FatalLevel, log.ErrorLevel,
		})
		if err != nil {
			log.Error(err)
			return
		}
		log.AddHook(hook)
	} else {
		log.Debug("dsn empty")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	log.Info("connecting to ", u.String())

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
				log.Error("read error:", err)
				return
			}
			jsonData, err := unzip(data)
			text := string(jsonData)
			if strings.Contains(text, "ping") {
				rspText := strings.Replace(text, "ping", "pong", 1)
				if err = c.WriteMessage(websocket.TextMessage, []byte(rspText)); err != nil {
					log.Error("WriteMessage: ", err)
				}
			} else {
				overview := &marketOverview{}
				if err = json.Unmarshal(jsonData, overview); err != nil {
					log.Error("json Unmarshal: ", err)
				}
				log.Info("receive data ", overview.Ts, len(overview.Data))
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

	for {
		select {
		case _ = <-ticker.C:
			usdtCoins := []string{"btc", "bch", "eth", "etc", "ltc", "eos", "xrp", "omg", "zec", "neo", "dash"}
			btcCoins := []string{"bch", "xrp", "eth", "ltc", "dash", "eos", "etc", "omg", "zec"}

			lock.RLock()
			usdtText := ""
			for _, coin := range usdtCoins {
				usdtText += getCoinText("usdt", coin, "%.2f")
			}

			btcText := ""
			for _, coin := range btcCoins {
				btcText += getCoinText("btc", coin, "%.6f")
			}

			lock.RUnlock()

			usdtText = "```\n" + usdtText + "\n```"
			usdtText = "*USDT*\n" + usdtText

			btcText = "```\n" + btcText + "\n```"
			btcText = "*BTC*\n" + btcText
			txt := usdtText + btcText

			if *tg {
				sendTG(txt)
				log.Info("send telegram")
			} else {
				log.Info("send debug output")
				sendDebug(txt)
			}

		case <-interrupt:
			log.Debug("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error("write close: ", err)
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
