package waterpump

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	BrokerUrl      string `json:"broker_url"`
	BrokerPort     int    `json:"broker_port"`
	BrokerUsername string `json:"broker_username"`
	BrokerPassword string `json:"broker_password"`

	DeviceId           string `json:"device_id"`
	DeviceName         string `json:"device_name"`
	DeviceModel        string `json:"device_model"`
	DeviceManufacturer string `json:"device_manufacturer"`
}

type ConfigurationLoader struct {
	Path string
}

func (l ConfigurationLoader) Load() (Configuration, error) {
	data, err := os.ReadFile(l.Path)
	if err != nil {
		return Configuration{}, err
	}
	var cfg Configuration
	return cfg, json.Unmarshal(data, &cfg)
}
