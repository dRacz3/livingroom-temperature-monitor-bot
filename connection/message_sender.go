package connection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type WebhookMessageSender struct {
	WebhookUrl string
}

type Payload struct {
	Text string `json:"text"`
}

func (s *WebhookMessageSender) SendMessage(text string) error {

	data := Payload{
		Text: text,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", s.WebhookUrl, body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()
	return nil

}
