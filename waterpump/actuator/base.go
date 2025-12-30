package actuator

import (
	"github.com/danilopeixoto/waterpump"
)

type BaseActuator struct {
	commands <-chan waterpump.ActuatorCommand
}

func (b *BaseActuator) SetCommands(ch <-chan waterpump.ActuatorCommand) {
	b.commands = ch
}
