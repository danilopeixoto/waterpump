package sensor

import (
	"github.com/danilopeixoto/waterpump"
)

type BaseSensor struct {
	telemetry chan<- waterpump.SensorTelemetry
}

func (b *BaseSensor) SetTelemetry(ch chan<- waterpump.SensorTelemetry) {
	b.telemetry = ch
}
