package sensor

import (
	"github.com/danilopeixoto/waterpump"
)

type SensorBase struct {
	telemetry chan<- waterpump.SensorTelemetry
}

func (s *SensorBase) SetTelemetry(ch chan<- waterpump.SensorTelemetry) {
	s.telemetry = ch
}
