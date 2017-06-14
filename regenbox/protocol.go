package regenbox

// see firmware/firmware.ino

const (
	Ping     byte = 0xa0
	StopByte byte = 0xff
)

const (
	ReadA0 byte = 0x00 | iota
	ReadVoltage
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
	ModeIdle byte = 0x50 | iota
	ModeCharge
	ModeDischarge
)
