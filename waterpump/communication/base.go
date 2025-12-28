package communication

import (
	"github.com/danilopeixoto/waterpump"
)

type CommunicationBase struct {
	telemetry <-chan waterpump.SensorTelemetry
	commands  chan<- waterpump.ActuatorCommand
}

func (c *CommunicationBase) SetTelemetry(ch <-chan waterpump.SensorTelemetry) {
	c.telemetry = ch
}

func (c *CommunicationBase) SetCommands(ch chan<- waterpump.ActuatorCommand) {
	c.commands = ch
}
