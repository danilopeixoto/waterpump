package waterpump

import "context"

type Runnable interface {
	Initialize(ctx context.Context) error
	Run(ctx context.Context)
	Deinitialize() error
}

type DiscoveryEntity struct {
	Entity  string
	Kind    string
	Options map[string]any
}

type Discoverable interface {
	DiscoveryEntities() []DiscoveryEntity
}

type Sensor interface {
	Runnable
	Discoverable
	SetTelemetry(chan<- SensorTelemetry)
}

type Actuator interface {
	Runnable
	Discoverable
	SetCommands(<-chan ActuatorCommand)
}

type Communication interface {
	Runnable
	SetTelemetry(chan<- SensorTelemetry)
	SetCommands(<-chan ActuatorCommand)
}

type SensorTelemetry struct {
	Entity string
	Value  any
}

type ActuatorCommand struct {
	Entity string
	Value  any
}
