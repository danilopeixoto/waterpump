package actuator

import (
	"context"

	"github.com/danilopeixoto/waterpump"
	"github.com/stianeikeland/go-rpio/v4"
)

type LedActuator struct {
	ActuatorBase
	Entity string
	Pin    rpio.Pin
}

func (l *LedActuator) Initialize(ctx context.Context) error {
	if err := rpio.Open(); err != nil {
		return err
	}
	l.Pin.Output()
	l.Pin.Low()
	return nil
}

func (l *LedActuator) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-l.commands:
			if cmd.Entity != l.Entity {
				continue
			}
			if cmd.Value == "ON" {
				l.Pin.High()
			} else {
				l.Pin.Low()
			}
		}
	}
}

func (l *LedActuator) Deinitialize() error {
	l.Pin.Low()
	rpio.Close()
	return nil
}

func (l *LedActuator) DiscoveryEntities() []waterpump.DiscoveryEntity {
	return []waterpump.DiscoveryEntity{
		{
			Entity: l.Entity,
			Kind:   "switch",
			Options: map[string]any{
				"payload_on":  "ON",
				"payload_off": "OFF",
			},
		},
	}
}
