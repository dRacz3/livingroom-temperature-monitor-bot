package connection

import (
	"fmt"
	"os/user"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MessageFloatMessageProcessor struct {
	LastMessage     string
	LastMessageTime time.Time
	ForwardChannel  chan float64
}

func (m *MessageFloatMessageProcessor) ProcessMessage(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	temp, err := strconv.ParseFloat((string(msg.Payload())), 64)
	if err != nil {
		m.ForwardChannel <- 0.0
	}
	m.ForwardChannel <- temp
}

type MqttMessageReceiver struct {
	channels []string
	broker   string
	port     int
	client   mqtt.Client
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic %s", topic)
}
func NewMqttMessageReceiver(topics []string, broker string, port int, messagePubHandler mqtt.MessageHandler) *MqttMessageReceiver {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))

	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	opts.SetClientID("go_mqtt_client_" + user.Name + "_" + time.Now().String())

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		fmt.Println("Connected")
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		fmt.Printf("Connect lost: %v", err)
	}

	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for _, channel := range topics {
		sub(client, channel)
	}

	return &MqttMessageReceiver{topics, broker, port, nil}
}
