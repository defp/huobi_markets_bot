package main

import (
	"net/url"
	"bytes"
	"net/http"

	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func sendSendmessage(text string) {
	params := url.Values{}
	params.Set("chat_id", "@huobi_pro_price")
	params.Set("text", text)
	params.Set("disable_notification", "true")
	body := bytes.NewBufferString(params.Encode())

	// Create client
	client := &http.Client{}

	// Create request
	url := "https://api.telegram.org/bot"+ *tgToken + "/sendMessage"
	req, err := http.NewRequest("POST", url, body)
	if err!= nil {
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

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display Results
	log.Debug("response Status : ", resp.Status)
	log.Debug("response Headers : ", resp.Header)
	log.Debug("response Body : ", string(respBody))
}

