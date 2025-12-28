package waterpump

import (
	"context"
	"sync"
)

type Device struct {
	Id         string
	Components []Runnable

	telemetry chan SensorTelemetry
	commands  chan ActuatorCommand

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (d *Device) Initialize(ctx context.Context) error {
	d.telemetry = make(chan SensorTelemetry, 32)
	d.commands = make(chan ActuatorCommand, 32)

	for _, c := range d.Components {
		if s, ok := c.(Sensor); ok {
			s.SetTelemetry(d.telemetry)
		}
		if a, ok := c.(Actuator); ok {
			a.SetCommands(d.commands)
		}
		if comm, ok := c.(Communication); ok {
			comm.SetTelemetry(d.telemetry)
			comm.SetCommands(d.commands)
		}
	}

	d.ctx, d.cancel = context.WithCancel(ctx)

	for _, c := range d.Components {
		if err := c.Initialize(d.ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) Run() {
	for _, c := range d.Components {
		d.wg.Add(1)
		go func(r Runnable) {
			defer d.wg.Done()
			r.Run(d.ctx)
		}(c)
	}
}

func (d *Device) Deinitialize() {
	d.cancel()
	d.wg.Wait()
	for _, c := range d.Components {
		_ = c.Deinitialize()
	}
}
