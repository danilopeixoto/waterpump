package actuator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/danilopeixoto/waterpump"
	"github.com/simonvetter/modbus"
)

type MebayActuator struct {
	BaseActuator
	DeviceId           string
	ControllerPortName string
	ControllerUnitId   uint8
	client             *modbus.ModbusClient
}

func (m *MebayActuator) Initialize(ctx context.Context) error {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:      fmt.Sprintf("rtu://%s", m.ControllerPortName),
		Speed:    19200,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 1,
		Timeout:  1 * time.Second,
	})

	if err != nil {
		return fmt.Errorf("failed to create Modbus client: %w", err)
	}

	err = client.Open()
	if err != nil {
		return fmt.Errorf("failed to open Modbus connection: %w", err)
	}

	m.client = client
	return nil
}

func (m *MebayActuator) Run(ctx context.Context) {
	var autoCommandMap = map[string]uint16{
		"On":  0x2222,
		"Off": 0x3333,
	}
	var powerCommandMap = map[string]uint16{
		"On":  0x4444,
		"Off": 0x1111,
	}

	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-m.commands:
			if cmd.Entity != "power" && cmd.Entity != "auto" {
				continue
			}

			err := m.client.SetUnitId(m.ControllerUnitId)
			if err != nil {
				log.Printf("Error setting unit ID: %v", err)
				continue
			}

			var value uint16
			var ok bool

			if cmd.Entity == "auto" {
				value, ok = autoCommandMap[cmd.Value.(string)]
			} else {
				value, ok = powerCommandMap[cmd.Value.(string)]
			}

			if !ok {
				log.Printf("Unknown command value: %v", cmd.Value)
				continue
			}

			err = m.client.WriteRegister(0x2001, value)
			if err != nil {
				log.Printf("Error writing to actuator: %v", err)
				continue
			}
		}
	}
}

func (m *MebayActuator) Deinitialize() error {
	if m.client != nil {
		return m.client.Close()
	}

	return nil
}

func (m *MebayActuator) DiscoveryEntities() []waterpump.DiscoveryEntity {
	base := "device/" + m.DeviceId
	return []waterpump.DiscoveryEntity{
		{
			Entity: "auto",
			Kind:   "switch",
			Options: map[string]any{
				"payload_on":  "On",
				"payload_off": "Off",
			},
		},
		{
			Entity: "power",
			Kind:   "switch",
			Options: map[string]any{
				"availability_topic": base + "/gear_status/state",
				"payload_available":  "Auto",
				"payload_on":         "On",
				"payload_off":        "Off",
			},
		},
	}
}
