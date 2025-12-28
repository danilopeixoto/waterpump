package communication

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/danilopeixoto/waterpump"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttCommunication struct {
	CommunicationBase

	BrokerUrl      string
	BrokerPort     int
	BrokerUsername string
	BrokerPassword string

	DeviceId           string
	DeviceName         string
	DeviceModel        string
	DeviceManufacturer string

	DiscoveryProviders []waterpump.Discoverable
	client             mqtt.Client
}

func (m *MqttCommunication) Initialize(ctx context.Context) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", m.BrokerUrl, m.BrokerPort))
	opts.SetClientID("device_" + m.DeviceId)
	opts.SetCleanSession(true)

	base := "device/" + m.DeviceId
	opts.SetWill(base+"/status", "offline", 1, true)

	if m.BrokerUsername != "" {
		opts.SetUsername(m.BrokerUsername)
	}
	if m.BrokerPassword != "" {
		opts.SetPassword(m.BrokerPassword)
	}

	deviceInfo := map[string]any{
		"identifiers":  []string{m.DeviceId},
		"name":         m.DeviceName,
		"model":        m.DeviceModel,
		"manufacturer": m.DeviceManufacturer,
	}

	opts.OnConnect = func(c mqtt.Client) {
		base := "device/" + m.DeviceId
		c.Subscribe(base+"/+/set", 1, nil)
		c.Publish(base+"/status", 1, true, "online")

		for _, p := range m.DiscoveryProviders {
			for _, e := range p.DiscoveryEntities() {
				payload := map[string]any{
					"unique_id":             m.DeviceId + "_" + e.Entity,
					"name":                  e.Entity,
					"availability_topic":    base + "/status",
					"payload_available":     "online",
					"payload_not_available": "offline",
					"state_topic":           base + "/" + e.Entity + "/state",
					"device":                deviceInfo,
				}

				controllableKinds := map[string]bool{
					"switch":  true,
					"light":   true,
					"cover":   true,
					"fan":     true,
					"climate": true,
					"lock":    true,
				}

				if controllableKinds[e.Kind] {
					payload["command_topic"] = base + "/" + e.Entity + "/set"
				}

				for k, v := range e.Options {
					payload[k] = v
				}
				c.Publish(
					fmt.Sprintf("homeassistant/%s/%s/%s/config", e.Kind, m.DeviceId, e.Entity),
					1, true,
					mustJSON(payload),
				)
			}
		}
	}

	opts.SetDefaultPublishHandler(func(c mqtt.Client, msg mqtt.Message) {
		entity := topicParts(msg.Topic())[2]
		value := string(msg.Payload())

		m.commands <- waterpump.ActuatorCommand{
			Entity: entity,
			Value:  value,
		}

		c.Publish(
			base+"/"+entity+"/state",
			1, true,
			value,
		)
	})

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *MqttCommunication) Run(ctx context.Context) {
	base := "device/" + m.DeviceId
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-m.telemetry:
			m.client.Publish(
				base+"/"+t.Entity+"/state",
				1, true,
				fmt.Sprintf("%v", t.Value),
			)
		}
	}
}

func (m *MqttCommunication) Deinitialize() error {
	base := "device/" + m.DeviceId
	m.client.Publish(base+"/status", 1, true, "offline")
	m.client.Disconnect(250)

	return nil
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func topicParts(topic string) []string {
	var parts []string
	start := 0
	for i := range topic {
		if topic[i] == '/' {
			parts = append(parts, topic[start:i])
			start = i + 1
		}
	}
	return append(parts, topic[start:])
}
