package actuator

import (
	"github.com/danilopeixoto/waterpump"
)

type ActuatorBase struct {
	commands <-chan waterpump.ActuatorCommand
}

func (a *ActuatorBase) SetCommands(ch <-chan waterpump.ActuatorCommand) {
	a.commands = ch
}
