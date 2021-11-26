package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

type TemperatureSensorStatus struct {
	is_available       bool
	temperature        float64
	last_status_change time.Time
}

type TemperatureSensorReading struct {
	status      bool
	temperature float64
}

func (t *TemperatureSensorStatus) update(status bool, temp float64) {
	t.is_available = status
	t.temperature = temp
	t.last_status_change = time.Now()
}

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

func main() {
	fmt.Println("Starting thermostat monitor")
	args := parse_args()

	channels := []string{args.channel_to_watch}

	temperature_channel := make(chan float64)
	ch := make(chan TemperatureSensorReading)
	slack_client := WebhookMessageSender{args.webhook_url}
	message_processor := MessageFloatMessageProcessor{"", time.Now(), temperature_channel}

	newMqttMessageReceiver(channels, args.broker_host, args.broker_port, message_processor.process_message)

	thermostat_status := TemperatureSensorStatus{
		is_available:       false,
		temperature:        0.0,
		last_status_change: time.Now(),
	}

	last_reported_status := TemperatureSensorStatus{
		is_available:       false,
		temperature:        0.0,
		last_status_change: time.Now(),
	}
	for { // loop forever
		go func() {
			select {
			case ret := <-temperature_channel:
				if ret > 0 {
					ch <- TemperatureSensorReading{true, ret}
				}
			case <-time.After(time.Second * 60): // 60 second timeout on the channel wait
				fmt.Println("Timeout...")
				ch <- TemperatureSensorReading{false, 0.0}
			}
		}()

		res := <-ch
		if res.status {

			last_message_was_long_time_ago := time.Until(last_reported_status.last_status_change) < -1*time.Hour

			if !thermostat_status.is_available || last_message_was_long_time_ago {
				slack_client.send_message(fmt.Sprintf("Living Room/Temperature: %f", res.temperature))
				thermostat_status.update(true, res.temperature)
				last_reported_status.update(true, res.temperature)
			}

		} else {
			if thermostat_status.is_available {
				slack_client.send_message("MIA Temperature sensor")
				thermostat_status.update(false, 0.0)
				println("No valid temperature")
			}
		}
	}
}
