package main

import (
	"flag"
	"fmt"
	"log"
	"mqtt_sentry/connection"
	"mqtt_sentry/sensor"
	"time"
)

type ParsedArgs struct {
	broker_host      string
	broker_port      int
	channel_to_watch string
	webhook_url      string
}

func parse_args() ParsedArgs {
	// variables declaration
	var broker_port int
	var broker_host string
	var channel_to_watch string
	var webhook_url string

	// flags declaration using flag package
	flag.IntVar(&broker_port, "port", 1883, "MQTT Broker port")
	flag.StringVar(&broker_host, "host", "localhost", "MQTT Broker host")
	flag.StringVar(&channel_to_watch, "channel", "living_room/temperature", "MQTT channel to watch")
	flag.StringVar(&webhook_url, "webhook", "", "Slack webhook url")
	flag.Parse()

	parsed := ParsedArgs{broker_host, broker_port, channel_to_watch, webhook_url}
	log.Println("Parsed args: ", parsed)
	return parsed
}

func gracefulExit() {
	failure := recover()
	if failure != nil {
		fmt.Printf("Failure: %#v", failure)
	}
}

func main() {
	defer gracefulExit()

	fmt.Println("Starting thermostat monitor")
	args := parse_args()

	channels := []string{args.channel_to_watch}

	temperature_channel := make(chan float64)
	ch := make(chan sensor.TemperatureSensorReading)
	slack_client := connection.WebhookMessageSender{WebhookUrl: args.webhook_url}
	message_processor := connection.MessageFloatMessageProcessor{LastMessage: "", LastMessageTime: time.Now(), ForwardChannel: temperature_channel}

	connection.NewMqttMessageReceiver(channels, args.broker_host, args.broker_port, message_processor.ProcessMessage)

	var thermostat_status sensor.TemperatureSensorStatus

	var last_reported_status sensor.TemperatureSensorStatus

	for { // loop forever
		go func() {
			select {
			case ret := <-temperature_channel:
				if ret > 0 {
					ch <- sensor.TemperatureSensorReading{Status: true, Temperature: ret}
				}
			case <-time.After(time.Second * 60): // 60 second timeout on the channel wait
				fmt.Println("Timeout...")
				ch <- sensor.TemperatureSensorReading{Status: false, Temperature: 0.0}
			}
		}()

		res := <-ch
		if res.Status {

			last_message_was_long_time_ago := time.Until(last_reported_status.LastStatusChange()) < -1*time.Hour

			if !thermostat_status.IsAvailable() || last_message_was_long_time_ago {
				slack_client.SendMessage(fmt.Sprintf("Living Room/Temperature: %f", res.Temperature))
				thermostat_status.Update(true, res.Temperature)
				last_reported_status.Update(true, res.Temperature)
			}

		} else {
			if thermostat_status.IsAvailable() {
				slack_client.SendMessage("MIA Temperature sensor")
				thermostat_status.Update(false, 0.0)
				println("No valid temperature")
			}
		}
	}
}
