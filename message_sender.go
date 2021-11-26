package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type WebhookMessageSender struct {
	webhook_url string
}

type Payload struct {
	Text string `json:"text"`
}

func (s *WebhookMessageSender) send_message(text string) {

	data := Payload{
		Text: text,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		println(err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", s.webhook_url, body)
	if err != nil {
		println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()
}
