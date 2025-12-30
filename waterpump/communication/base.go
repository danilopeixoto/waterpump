package communication

import (
	"github.com/danilopeixoto/waterpump"
)

type BaseCommunication struct {
	telemetry <-chan waterpump.SensorTelemetry
	commands  chan<- waterpump.ActuatorCommand
}

func (b *BaseCommunication) SetTelemetry(ch <-chan waterpump.SensorTelemetry) {
	b.telemetry = ch
}

func (b *BaseCommunication) SetCommands(ch chan<- waterpump.ActuatorCommand) {
	b.commands = ch
}
