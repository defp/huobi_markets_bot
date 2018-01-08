package main

import (
	"flag"
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

	"fmt"
	"github.com/evalphobia/logrus_sentry"
	log "github.com/sirupsen/logrus"
	"sync"
)

var addr = flag.String("addr", "api.huobi.pro", "http service address")
var tgToken = flag.String("tgToken", "", "telegram token")
var dsn = flag.String("dsn", "", "sentry dsn")
var second = flag.Int("second", 60, "telegram send drution")

var coinClosePrice = make(map[string]TickerData)
var lock sync.RWMutex

type TickerData struct {
	Open   float64
	Close  float64
	Low    float64
	High   float64
	Amount float64
	Count  int64
	Vol    float64
	Symbol string
	Range  float64
}

type MarketOverview struct {
	Ch     string
	Ts     int64
	Status string
	Data   []TickerData
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

	log.SetOutput(os.Stdout)
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
				overview := &MarketOverview{}
				if err = json.Unmarshal(jsonData, overview); err != nil {
					log.Error("json Unmarshal: ", err)
				}
				log.Info("receive data ", overview.Ts, len(overview.Data))
				lock.Lock()
				for _, data := range overview.Data {
					preClosePrice := coinClosePrice[data.Symbol].Close
					data.Range = data.Close - preClosePrice
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
			//c.WriteMessage(websocket.TextMessage, []byte(`{"req": "market.xrpusdt.trade.detail", "id": "id1"}`))
			//c.WriteMessage(websocket.TextMessage, []byte(`{"req": "market.xrpusdt.detail", "id": "id2"}`))

			coins := []string{"btc", "bch", "xrp", "eth", "ltc", "eos", "etc", "omg", "zec", "snt", "neo", "hsr", "dash", "qtum"}

			tgText := ""
			lock.RLock()
			for _, coin := range coins {
				tickerData := coinClosePrice[coin+"usdt"]
				coinName := padRight(strings.ToUpper(coin), 4, " ")

				closePrice := padRight(fmt.Sprintf("%.2f", tickerData.Close), 9, " ")

				var riseText string
				if tickerData.Range > 0 {
					upText := padRight("up", 4, " ")
					riseText = fmt.Sprintf(upText + " $%s", fmt.Sprintf("%.2f", tickerData.Range))
				} else {
					riseText = fmt.Sprintf("Down $%s", fmt.Sprintf("%.2f", tickerData.Range))
				}

				tgText = tgText + fmt.Sprintf("%s $%s %s\n", coinName, closePrice, riseText)
			}
			lock.RUnlock()
			//sendTG(tgText)
			sendDebug(tgText)

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
