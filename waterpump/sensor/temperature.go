package sensor

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/danilopeixoto/waterpump"
)

type TemperatureSensor struct {
	SensorBase
}

func (t *TemperatureSensor) Initialize(ctx context.Context) error { return nil }

func (t *TemperatureSensor) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.telemetry <- waterpump.SensorTelemetry{
				Entity: "temperature",
				Value:  fmt.Sprintf("%.1f", 20+rand.Float64()*10),
			}
		}
	}
}

func (t *TemperatureSensor) Deinitialize() error { return nil }

func (t *TemperatureSensor) DiscoveryEntities() []waterpump.DiscoveryEntity {
	return []waterpump.DiscoveryEntity{
		{
			Entity: "temperature",
			Kind:   "sensor",
			Options: map[string]any{
				"unit_of_measurement": "°C",
				"device_class":        "temperature",
			},
		},
	}
}
