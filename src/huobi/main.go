package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"time"
)

var addr = flag.String("addr", "api.huobi.pro", "http service address")
var tgToken = flag.String("tgToken", "", "telegram token")
var second = flag.Int("second", 1, "telegram send drution")

var coinClosePrice = make(map[string]tickerData)
var lock sync.RWMutex
var lastSendText = ""

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go ginWeb()
	go huobiWebsocket()

	ticker := time.NewTicker(time.Duration(*second) * time.Second)
	defer ticker.Stop()

	usdtCoins := []string{
		"btc", "bch", "eth", "etc", "ltc", "eos", "xrp", "omg", "zec", "neo", "dash",
		"ht", "qtum", "hsr", "dta", "let", "snt", "cvc", "smt", "ven", "elf", "xem"}

	for {
		select {
		case _ = <-ticker.C:
			lock.RLock()
			text := ""
			for _, coin := range usdtCoins {
				text += getCoinText("usdt", coin, "%.2f")
			}

			lock.RUnlock()
			lastSendText = text
			if *tgToken != "" {
				sendTG("*USDT*\n```\n" + text + "\n```")
			}
		case <-interrupt:
			return
		}
	}
}
