package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/danilopeixoto/waterpump"
	"github.com/danilopeixoto/waterpump/actuator"
	"github.com/danilopeixoto/waterpump/communication"
	"github.com/danilopeixoto/waterpump/sensor"
)

func main() {
	loader := waterpump.ConfigurationLoader{Path: "/data/options.json"}
	cfg, err := loader.Load()
	if err != nil {
		log.Fatal(err)
	}

	sensor := &sensor.MebaySensor{
		DeviceId:            cfg.DeviceId,
		ControllerPortName:  cfg.ControllerPortName,
		ControllerUnitId:    uint8(cfg.ControllerUnitId),
		PollIntervalSeconds: 2,
	}

	actuator := &actuator.MebayActuator{
		DeviceId:           cfg.DeviceId,
		ControllerPortName: cfg.ControllerPortName,
		ControllerUnitId:   uint8(cfg.ControllerUnitId),
	}

	mqtt := &communication.MqttCommunication{
		BrokerUrl:          cfg.BrokerUrl,
		BrokerPort:         cfg.BrokerPort,
		BrokerUsername:     cfg.BrokerUsername,
		BrokerPassword:     cfg.BrokerPassword,
		DeviceId:           cfg.DeviceId,
		DeviceName:         cfg.DeviceName,
		DeviceModel:        cfg.DeviceModel,
		DeviceManufacturer: cfg.DeviceManufacturer,
		DiscoveryProviders: []waterpump.Discoverable{
			sensor,
			actuator,
		},
	}

	device := waterpump.Device{
		Id: cfg.DeviceId,
		Components: []waterpump.Runnable{
			sensor,
			actuator,
			mqtt,
		},
	}

	ctx := context.Background()
	if err := device.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	device.Run()
	waitForExit()
	device.Deinitialize()
}

func waitForExit() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
