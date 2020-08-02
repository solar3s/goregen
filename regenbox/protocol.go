package regenbox

// see firmware/firmware.ino

const (
	Ping     byte = 0xa0
	StopByte byte = 0xff
)

const (
	ReadA0 byte = 0x00 | iota
	ReadVoltage0
	ReadFirmware
	ReadVoltage1 byte = 0x05
	ReadVoltage2 byte = 0x06
	ReadVoltage3 byte = 0x07
)

const (
	LedOff byte = 0x10 | iota
	LedOn
	LedToggle
)

const (
	PinDischargeOff byte = 0x20 | iota
	PinDischargeOn
)

const (
	PinChargeOff byte = 0x30 | iota
	PinChargeOn
)

const (
	// beware of any changes here, voodoo was made in types_marshallers.go:89,98
	// to ensure both equality with ChargeState types and fancy marshallers
	ModeIdle      byte = 0x50
	ModeCharge         = 0x51
	ModeDischarge      = 0x52
	ModeChargeX4       = 0x53
)
