package main

import (
	"flag"
	"fmt"
	"log"
	"mqtt_sentry/connection"
	"mqtt_sentry/sensor"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
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

func gracefulExit(messageSender connection.WebhookMessageSender) {
	failure := recover()
	if failure != nil {
		fmt.Printf("Failure: %#v", failure)
		messageSender.SendMessage(fmt.Sprintf("MQTT Monitor exit: %#v", failure))
		fmt.Println("Exiting...")

	}
}

func mqttLoop() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	fmt.Println("Starting thermostat monitor")
	args := parse_args()

	channels := []string{args.channel_to_watch}

	temperatureChannel := make(chan float64)
	ch := make(chan sensor.TemperatureSensorReading)
	slackClient := connection.WebhookMessageSender{WebhookUrl: args.webhook_url}
	defer gracefulExit(slackClient)
	go func() {
		for sig := range c {
			slackClient.SendMessage(fmt.Sprintf("MQTT Monitor exit signal caught: %#v", sig))
			log.Fatal("Caused by: ", sig)
		}
	}()

	messageProcessor := connection.MessageFloatMessageProcessor{LastMessage: "", LastMessageTime: time.Now(), ForwardChannel: temperatureChannel}

	connection.NewMqttMessageReceiver(channels, args.broker_host, args.broker_port, messageProcessor.ProcessMessage)

	var thermostatStatus sensor.TemperatureSensorStatus

	var lastReportedStatus sensor.TemperatureSensorStatus

	for { // loop forever
		go func() {
			select {
			case ret := <-temperatureChannel:
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

			last_message_was_long_time_ago := time.Until(lastReportedStatus.LastStatusChange()) < -1*time.Hour

			if !thermostatStatus.IsAvailable() || last_message_was_long_time_ago {
				slackClient.SendMessage(fmt.Sprintf("Living Room/Temperature: %f", res.Temperature))
				thermostatStatus.Update(true, res.Temperature)
				lastReportedStatus.Update(true, res.Temperature)
			}

		} else {
			if thermostatStatus.IsAvailable() {
				slackClient.SendMessage("MIA Temperature sensor")
				thermostatStatus.Update(false, 0.0)
				fmt.Println("No valid temperature received")
			}
		}
	}
}

func loopServerForStatus() {
	app := fiber.New()

	app.Get("/metrics", monitor.New(monitor.Config{Title: "Heart of Irongate"}))
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})
	log.Fatal(app.Listen(":3000"))

}

func main() {
	go mqttLoop()
	loopServerForStatus()
}
