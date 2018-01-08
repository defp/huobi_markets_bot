package main

import (
	"bytes"
	"net/http"
	"net/url"

	"io/ioutil"

	"fmt"
	log "github.com/sirupsen/logrus"
)

func sendTG(text string) {
	params := url.Values{}
	params.Set("chat_id", "@huobi_pro_price")
	params.Set("text", text)
	params.Set("disable_notification", "true")
	body := bytes.NewBufferString(params.Encode())

	// Create client
	client := &http.Client{}

	// Create request
	url := "https://api.telegram.org/bot" + *tgToken + "/sendMessage"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Error("NewRequest error: ", err)
		return
	}

	// Headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		log.Error("Failure : ", err)
		return
	}

	if resp.StatusCode != 200 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Error("request error", respBody)
	}
}

func sendDebug(text string) {
	fmt.Print(text)
}
