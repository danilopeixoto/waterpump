package sensor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/danilopeixoto/waterpump"
	"github.com/simonvetter/modbus"
)

type MebaySensor struct {
	BaseSensor
	DeviceId            string
	ControllerPortName  string
	ControllerUnitId    uint8
	PollIntervalSeconds int
	client              *modbus.ModbusClient
}

type MebaySensorData struct {
	Rpm                float32
	WaterTemperature   float32
	OilPressure        float32
	CurrentRunningTime uint64
	TotalRunningTime   uint64
	BatteryVoltage     float32
	ChargingVoltage    float32
	GearStatus         string
	RunningStatus      string
	AlarmCode          string
	WarningCodes       []string
}

func (m *MebaySensor) Initialize(ctx context.Context) error {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:      fmt.Sprintf("rtu://%s", m.ControllerPortName),
		Speed:    19200,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 1,
		Timeout:  1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create Modbus client: %w", err)
	}

	err = client.Open()
	if err != nil {
		return fmt.Errorf("failed to open Modbus connection: %w", err)
	}

	m.client = client
	return nil
}

func (m *MebaySensor) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(m.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := m.client.SetUnitId(m.ControllerUnitId)
			if err != nil {
				log.Printf("Error setting unit ID: %v", err)
				continue
			}

			data, err := m.readSensorData()
			if err != nil {
				log.Printf("Error reading from sensor: %v", err)
				continue
			}

			m.sendTelemetry(data)
		}
	}
}

func (m *MebaySensor) readSensorData() (*MebaySensorData, error) {
	var gearStatusMap = map[uint16]string{
		0x0033: "Stop",
		0x0066: "Local",
		0x0099: "Auto",
		0x00CC: "Remote",
	}

	var runningStatusMap = map[uint16]string{
		0x0000: "Stop Idle Speed",
		0x0001: "Under stop",
		0x0002: "Waiting",
		0x0003: "Crank Cancel",
		0x0004: "Crank Interval",
		0x0005: "Alarm Reset",
		0x0006: "Standby",
		0x0007: "Crank Delay",
		0x0008: "Pre-Heat",
		0x0009: "Pre-Oil Supply",
		0x000A: "Suction Pump Starting",
		0x000B: "Suction Pump Start-Up Interval",
		0x000C: "Waiting for Suction Pump Pressure",
		0x000D: "Crank Ready",
		0x000E: "In Crank",
		0x000F: "Oil Pressure Delay",
		0x0010: "Safety Delay",
		0x0011: "Idle Speed",
		0x0012: "Potentiometer Speed Mode",
		0x0013: "Speed-Up",
		0x0014: "Boosting Operation",
		0x0015: "Pressure Building and Boosting",
		0x0016: "High-Speed Warming",
		0x0017: "Rated Running",
		0x0018: "Return Delay",
		0x0019: "Return to Primary",
		0x001A: "Cooling Running",
		0x001B: "Return to Power Generation",
		0x001C: "Alarm After Waiting for Heat Dissipation",
		0x001D: "Switchover",
		0x001E: "Setting in Progress",
	}

	var alarmCodeMap = map[uint16]string{
		0x0000: "None",
		0x0001: "RPM Signal Lost",
		0x0002: "Over Speed",
		0x0003: "Under Speed",
		0x0004: "Custom Sensor Open Circuit",
		0x0005: "Speed Potentiometer Open Circuit",
		0x0006: "Temperature Open",
		0x0007: "Oil Pressure Open",
		0x0008: "Liquid Level Open",
		0x0009: "Fuel Level Open",
		0x000A: "Network Pressure Open",
		0x000B: "Water Inlet Pressure Open",
		0x000C: "Water Outlet Pressure Open",
		0x000D: "Torque Converter Oil Temperature Open Circuit",
		0x000E: "Water Level Open",
		0x000F: "Water Pressure Open",
		0x0010: "Coolant Temperature Open",
		0x0011: "Exhaust Gas Temperature Open",
		0x0012: "Exhaust Pressure Open",
		0x0013: "Intake Pressure Open",
		0x0014: "Primary Pressure Open",
		0x0015: "Secondary Pressure Open",
		0x0016: "System Pressure Open",
		0x0017: "Flow Open",
		0x0018: "Custom Sensor 1 High",
		0x0019: "Speed Potentiometer High",
		0x001A: "Temperature High",
		0x001B: "Oil Pressure High",
		0x001C: "Liquid Level High",
		0x001D: "Fuel Level High",
		0x001E: "Network Pressure High",
		0x001F: "Inlet Water Pressure High",
		0x0020: "Water Outlet Pressure High",
		0x0021: "Torque Converter Oil Temperature High",
		0x0022: "Water Level High",
		0x0023: "High Water Pressure",
		0x0024: "Coolant Temperature High",
		0x0025: "Exhaust Gas Temperature High",
		0x0026: "Exhaust Pressure High",
		0x0027: "Intake Pressure High",
		0x0028: "Primary Pressure High",
		0x0029: "Secondary Pressure High",
		0x002A: "System Pressure High",
		0x002B: "High Flow Rate",
		0x002C: "Custom Sensor 1 High",
		0x002D: "Speed Potentiometer High",
		0x002E: "Temperature High",
		0x002F: "Oil Pressure High",
		0x0030: "Liquid Level High",
		0x0031: "Fuel Level High",
		0x0032: "Network Pressure High",
		0x0033: "Inlet Water Pressure High",
		0x0034: "Water Outlet Pressure High",
		0x0035: "Torque Converter Oil Temperature High",
		0x0036: "Water Level High",
		0x0037: "High Water Pressure",
		0x0038: "Coolant Temperature High",
		0x0039: "Exhaust Gas Temperature High",
		0x003A: "Exhaust Pressure High",
		0x003B: "Intake Pressure High",
		0x003C: "Primary Pressure High",
		0x003D: "Secondary Pressure High",
		0x003E: "System Pressure High",
		0x003F: "High Flow Rate",
		0x0040: "Custom Sensor 1 Low",
		0x0041: "Speed Potentiometer Low",
		0x0042: "Temperature Low",
		0x0043: "Oil Pressure Low",
		0x0044: "Fluid Level Low",
		0x0045: "Fuel Level Low",
		0x0046: "Low Network Pressure",
		0x0047: "Inlet Water Pressure Low",
		0x0048: "Low Water Pressure",
		0x0049: "Torque Converter Oil Temperature Low",
		0x004A: "Water Level Low",
		0x004B: "Low Water Pressure",
		0x004C: "Coolant Temperature Low",
		0x004D: "Exhaust Temperature Low",
		0x004E: "Exhaust Pressure Low",
		0x004F: "Low Intake Pressure",
		0x0050: "Primary Pressure Low",
		0x0051: "Secondary Pressure Low",
		0x0052: "System Pressure Low",
		0x0053: "Low Flow Rate",
		0x0054: "Custom Sensor 1 Low",
		0x0055: "Speed Potentiometer Low",
		0x0056: "Temperature Low",
		0x0057: "Oil Pressure Low",
		0x0058: "Fluid Level Low",
		0x0059: "Fuel Level Low",
		0x005A: "Low Network Pressure",
		0x005B: "Inlet Water Pressure Low",
		0x005C: "Low Water Pressure",
		0x005D: "Torque Converter Oil Temperature Low",
		0x005E: "Water Level Low",
		0x005F: "Low Water Pressure",
		0x0060: "Coolant Temperature Low",
		0x0061: "Exhaust Temperature Low",
		0x0062: "Exhaust Pressure Low",
		0x0063: "Low Intake Pressure",
		0x0064: "Primary Pressure Low",
		0x0065: "Secondary Pressure Low",
		0x0066: "System Pressure Low",
		0x0067: "Low Flow Rate",
		0x0068: "Over Flow",
		0x0069: "Low Oil Pressure Alarm Switch",
		0x006A: "Low Oil Pressure Warning Switch",
		0x006B: "High Temperature Alarm Switch",
		0x006C: "High Temperature Warning Switch",
		0x006D: "Low Water Level Alarm Switch",
		0x006E: "Low Water Level Warning Switch",
		0x006F: "Fuel Level Low Alarm Switch",
		0x0070: "Fuel Level Low Warning Switch",
		0x0071: "Start Failure",
		0x0072: "Stop Failure Rpm",
		0x0073: "Stop Failure Oil Pressure",
		0x0074: "Stop Failure Oil Pressure Switch",
		0x0075: "Oil Filter Maintenance Countdown",
		0x0076: "Air Filter Maintenance Countdown",
		0x0077: "Fuel Filter Maintenance Countdown",
		0x0078: "Oil Filter Maintenance",
		0x0079: "Air Filter Maintenance",
		0x007A: "Fuel Filter Maintenance",
		0x007B: "ECU Alarm",
		0x007C: "ECU Comms Failure",
		0x007D: "Emergency Stop",
		0x007E: "485 Comms Failure",
		0x007F: "External Instant Alarm Switch",
		0x0080: "External Stop Alarm Switch",
		0x0081: "Suction Pump Start Failure",
		0x0082: "Suction Pump Failure",
		0x0083: "Blinds Abnormal",
		0x0084: "Oil Pressure Too Low",
		0x0085: "Temperature Too High",
		0x0086: "Sub Engine Emergency Stop",
		0x0087: "Custom 1 Input",
		0x0088: "Custom 2 Input",
		0x0089: "Custom 3 Input",
		0x008A: "Custom 4 Input",
		0x008B: "Custom 5 Input",
		0x008C: "Custom 6 Input",
		0x008D: "Custom Sensor 2 High",
		0x008E: "Custom Sensor 3 High",
		0x008F: "Custom Sensor 4 High",
		0x0090: "Custom Sensor 5 High",
		0x0091: "Custom Sensor 6 High",
		0x0092: "Custom Sensor 2 High",
		0x0093: "Custom Sensor 3 High",
		0x0094: "Custom Sensor 4 High",
		0x0095: "Custom Sensor 5 High",
		0x0096: "Custom Sensor 6 High",
		0x0097: "Custom Sensor 2 Low",
		0x0098: "Custom Sensor 3 Low",
		0x0099: "Custom Sensor 4 Low",
		0x009A: "Custom Sensor 5 Low",
		0x009B: "Custom Sensor 6 Low",
		0x009C: "Custom Sensor 2 Low",
		0x009D: "Custom Sensor 3 Low",
		0x009E: "Custom Sensor 4 Low",
		0x009F: "Custom Sensor 5 Low",
		0x00A0: "Custom Sensor 6 Low",
	}

	var warningCodeMap = map[uint16]string{
		0:  "Fuel Filter Maintenance Countdown",
		1:  "ECU Warning",
		2:  "ECU Communication Failure",
		3:  "Emergency Stop",
		4:  "485 Communication Failure",
		5:  "External Instant Alarm Switch",
		6:  "External Stop Warning Switch",
		7:  "Charger Charge Failure",
		8:  "Fuel Filling Failure",
		9:  "Custom 1 Input",
		10: "External Charger Charging Failure",
		11: "Custom 2 Input Warning",
		12: "Custom 3 Input Warning",
		13: "Custom 4 Input Warning",
		14: "Custom 5 Input Warning",
		15: "Custom 6 Input Warning",

		16: "Low Flow Rate",
		17: "Over Flow",
		18: "Low Oil Pressure Warning Switch",
		19: "Low Oil Pressure Warning Switch",
		20: "High Temperature Switch",
		21: "High Temperature Warning Switch",
		22: "Low Water Level Warning Switch",
		23: "Low Water Level Warning Switch",
		24: "Low Fuel Level Stop Switch",
		25: "Fuel Level Low Warning Switch",
		26: "Start Failure",
		27: "Stop Failure RPM",
		28: "Stop Failure Oil Pressure",
		29: "Stop Failure Oil Pressure Switch",
		30: "Oil Filter Service Countdown",
		31: "Air Filter Service Countdown",

		32: "Low Oil Pressure",
		33: "Low Fluid Level",
		34: "Fuel Level Low",
		35: "Low Network Pressure",
		36: "Low Inlet Pressure",
		37: "Low Water Pressure",
		38: "Low Torque Converter Oil Temperature",
		39: "Low Water Level",
		40: "Low Water Pressure",
		41: "Low Coolant Temperature",
		42: "Low Exhaust Temperature",
		43: "Low Exhaust Pressure",
		44: "Low Intake Pressure",
		45: "Low Primary Pressure",
		46: "Low Secondary Pressure",
		47: "Low System Pressure",

		48: "High Inlet Pressure",
		49: "High Water Pressure",
		50: "High Torque Converter Oil Temperature",
		51: "High Water Level",
		52: "High Water Pressure",
		53: "High Coolant Temperature",
		54: "High Exhaust Temperature",
		55: "High Exhaust Pressure",
		56: "High Intake Pressure",
		57: "High Primary Pressure",
		58: "High Secondary Pressure",
		59: "High System Pressure",
		60: "High Flow Rate",
		61: "Custom Sensor Low",
		62: "Speed Potentiometer Low",
		63: "Temperature Low",

		64: "Water Pressure Open Circuit",
		65: "Coolant Temperature Open",
		66: "Exhaust Temperature Open",
		67: "Exhaust Pressure Open Circuit",
		68: "Intake Pressure Open Circuit",
		69: "Primary Pressure Open",
		70: "Secondary Pressure Open",
		71: "System Pressure Open",
		72: "Flow Open",
		73: "Custom Sensor High",
		74: "Speed Potentiometer High",
		75: "Temperature High",
		76: "Oil Pressure High",
		77: "Fluid Level High",
		78: "Fuel Level High",
		79: "High Network Pressure",

		80: "Loss of Tacho Signal",
		81: "Overspeed",
		82: "Under Speed",
		83: "Battery Overvoltage",
		84: "Low Battery Voltage",
		85: "Custom Sensor Open Circuit",
		86: "Open Speed Potentiometer",
		87: "Temperature Open",
		88: "Oil Pressure Open",
		89: "Liquid Level Open",
		90: "Fuel Level Open",
		91: "Network Pressure Open",
		92: "Water Inlet Pressure Open Circuit",
		93: "Water Outlet Pressure Open Circuit",
		94: "Torque Converter Oil Temperature Open Circuit",
		95: "Water Level Open",
	}

	data := &MebaySensorData{}

	rpm, err := m.client.ReadRegister(0x1000, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read RPM: %w", err)
	}
	data.Rpm = float32(rpm)

	waterTemperature, err := m.client.ReadRegister(0x1025, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read water temperature: %w", err)
	}
	data.WaterTemperature = float32(waterTemperature)

	oilPressure, err := m.client.ReadRegister(0x1024, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read oil pressure: %w", err)
	}
	data.OilPressure = float32(oilPressure)

	currentRunningTimeRegs, err := m.client.ReadRegisters(0x1012, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read current running time: %w", err)
	}
	data.CurrentRunningTime = (uint64(currentRunningTimeRegs[0])<<16 | uint64(currentRunningTimeRegs[1])) * 360

	totalRunningTimeRegs, err := m.client.ReadRegisters(0x1015, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read total running time: %w", err)
	}
	data.TotalRunningTime = (uint64(totalRunningTimeRegs[0])<<16 | uint64(totalRunningTimeRegs[1])) * 360

	batteryVoltage, err := m.client.ReadRegister(0x1001, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read battery voltage: %w", err)
	}
	data.BatteryVoltage = float32(batteryVoltage) * 0.1

	chargingVoltage, err := m.client.ReadRegister(0x1002, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read charging voltage: %w", err)
	}
	data.ChargingVoltage = float32(chargingVoltage) * 0.1

	gearStatus, err := m.client.ReadRegister(0x1017, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read gear status: %w", err)
	}
	if msg, ok := gearStatusMap[gearStatus]; ok {
		data.GearStatus = msg
	} else {
		data.GearStatus = "Unknown"
	}

	runningStatus, err := m.client.ReadRegister(0x1018, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read running status: %w", err)
	}
	if msg, ok := runningStatusMap[runningStatus]; ok {
		data.RunningStatus = msg
	} else {
		data.RunningStatus = "Unknown"
	}

	alarmCode, err := m.client.ReadRegister(0x101A, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read alarm code: %w", err)
	}
	if msg, ok := alarmCodeMap[alarmCode]; ok {
		data.AlarmCode = msg
	} else {
		data.AlarmCode = "Unknown"
	}

	warningCodeRegs, err := m.client.ReadRegisters(0x101B, 6, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("failed to read warning codes: %w", err)
	}

	data.WarningCodes = []string{}
	for regIndex, regValue := range warningCodeRegs {
		if regValue == 0 {
			continue
		}

		codeBase := uint16(regIndex) * 16
		for bit := range uint16(16) {
			if regValue&(1<<bit) != 0 {
				if msg, ok := warningCodeMap[codeBase+bit]; ok {
					data.WarningCodes = append(data.WarningCodes, msg)
				}
			}
		}
	}

	return data, nil
}

func (m *MebaySensor) sendTelemetry(data *MebaySensorData) {
	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "rpm",
		Value:  data.Rpm,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "water_temperature",
		Value:  data.WaterTemperature,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "oil_pressure",
		Value:  data.OilPressure,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "current_running_time",
		Value:  data.CurrentRunningTime,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "total_running_time",
		Value:  data.TotalRunningTime,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "battery_voltage",
		Value:  data.BatteryVoltage,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "charging_voltage",
		Value:  data.ChargingVoltage,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "gear_status",
		Value:  data.GearStatus,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "running_status",
		Value:  data.RunningStatus,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "alarm_code",
		Value:  data.AlarmCode,
	}

	m.telemetry <- waterpump.SensorTelemetry{
		Entity: "warning_codes",
		Value: map[string]any{
			"items": data.WarningCodes,
			"count": len(data.WarningCodes),
		},
	}
}

func (m *MebaySensor) Deinitialize() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

func (m *MebaySensor) DiscoveryEntities() []waterpump.DiscoveryEntity {
	base := "device/" + m.DeviceId
	return []waterpump.DiscoveryEntity{
		{
			Entity: "rpm",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class":        "speed",
				"unit_of_measurement": "rpm",
			},
		},
		{
			Entity: "water_temperature",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class":        "temperature",
				"unit_of_measurement": "°C",
			},
		},
		{
			Entity: "oil_pressure",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class":        "pressure",
				"unit_of_measurement": "kPa",
			},
		},
		{
			Entity: "current_running_time",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class":        "duration",
				"unit_of_measurement": "s",
			},
		},
		{
			Entity: "total_running_time",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class":        "duration",
				"unit_of_measurement": "s",
			},
		},
		{
			Entity: "battery_voltage",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class":        "voltage",
				"unit_of_measurement": "V",
			},
		},
		{
			Entity: "charging_voltage",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class":        "voltage",
				"unit_of_measurement": "V",
			},
		},
		{
			Entity: "gear_status",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class": "enum",
			},
		},
		{
			Entity: "running_status",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class": "enum",
			},
		},
		{
			Entity: "alarm_code",
			Kind:   "sensor",
			Options: map[string]any{
				"device_class": "enum",
			},
		},
		{
			Entity: "warning_codes",
			Kind:   "sensor",
			Options: map[string]any{
				"json_attributes_topic":    base + "/warning_codes/state",
				"value_template":           "{{ value_json.count }}",
				"json_attributes_template": "{{ value_json | tojson }}",
			},
		},
	}
}
