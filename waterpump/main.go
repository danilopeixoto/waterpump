package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const OptionsFile = "/data/options.json"

type OptionsData struct {
	BrokerURL      string `json:"broker_url"`
	BrokerPort     int    `json:"broker_port"`
	BrokerUsername string `json:"broker_username"`
	BrokerPassword string `json:"broker_password"`
	DeviceID       string `json:"device_id"`
	DeviceName     string `json:"device_name"`
}

var (
	Options           OptionsData
	BaseTopic         string
	AvailabilityTopic string
	TemperatureTopic  string
	PowerSetTopic     string
	PowerStateTopic   string
)

func main() {
	loadOptions()
	buildTopics()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", Options.BrokerURL, Options.BrokerPort))
	opts.SetClientID("device_" + Options.DeviceID)

	if Options.BrokerUsername != "" {
		opts.SetUsername(Options.BrokerUsername)
	}
	if Options.BrokerPassword != "" {
		opts.SetPassword(Options.BrokerPassword)
	}

	opts.SetCleanSession(true)

	opts.SetWill(AvailabilityTopic, "offline", 1, true)

	opts.OnConnect = onConnect
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Println("Connection lost:", err)
	}
	opts.SetDefaultPublishHandler(onMessage)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			temp := 20 + rand.Float64()*10
			client.Publish(
				TemperatureTopic,
				1,
				true,
				fmt.Sprintf("%.1f", temp),
			)
		case <-sig:
			client.Publish(AvailabilityTopic, 1, true, "offline")
			client.Disconnect(250)
			return
		}
	}
}

func loadOptions() {
	data, err := os.ReadFile(OptionsFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(data, &Options); err != nil {
		log.Fatal(err)
	}
}

func buildTopics() {
	BaseTopic = "device/" + Options.DeviceID
	AvailabilityTopic = BaseTopic + "/status"
	TemperatureTopic = BaseTopic + "/temperature"
	PowerSetTopic = BaseTopic + "/power/set"
	PowerStateTopic = BaseTopic + "/power/state"
}

func onConnect(client mqtt.Client) {
	client.Publish(AvailabilityTopic, 1, true, "online")
	client.Subscribe(PowerSetTopic, 1, nil)
	publishDiscovery(client)
}

func onMessage(client mqtt.Client, msg mqtt.Message) {
	payload := string(msg.Payload())
	client.Publish(PowerStateTopic, 1, true, payload)
}

func publishDiscovery(client mqtt.Client) {
	deviceInfo := map[string]any{
		"identifiers":  []string{Options.DeviceID},
		"name":         Options.DeviceName,
		"model":        "Diesel",
		"manufacturer": "Tangerina",
	}

	tempConfig := map[string]any{
		"name":                  "Temperature",
		"state_topic":           TemperatureTopic,
		"device_class":          "temperature",
		"unit_of_measurement":   "°C",
		"availability_topic":    AvailabilityTopic,
		"payload_available":     "online",
		"payload_not_available": "offline",
		"unique_id":             Options.DeviceID + "_temperature",
		"device":                deviceInfo,
	}

	powerConfig := map[string]any{
		"name":                  "Power",
		"command_topic":         PowerSetTopic,
		"state_topic":           PowerStateTopic,
		"availability_topic":    AvailabilityTopic,
		"payload_on":            "ON",
		"payload_off":           "OFF",
		"payload_available":     "online",
		"payload_not_available": "offline",
		"unique_id":             Options.DeviceID + "_power",
		"device":                deviceInfo,
	}

	publishJSON(
		client,
		"homeassistant/sensor/"+Options.DeviceID+"/temperature/config",
		tempConfig,
	)
	publishJSON(
		client,
		"homeassistant/switch/"+Options.DeviceID+"/power/config",
		powerConfig,
	)
}

func publishJSON(client mqtt.Client, topic string, payload any) {
	data, _ := json.Marshal(payload)
	client.Publish(topic, 1, true, data)
}
