package ata

import "github.com/rnmz/gosmart/internal/transport"

const lbaDevice = 0x40

var (
	IdentifyDevice = transport.ATACommand{Command: 0xEC, Device: lbaDevice}

	SmartReadData = transport.ATACommand{
		Command: 0xB0, Features: 0xD0, LBAMid: 0x4F, LBAHigh: 0xC2, Device: lbaDevice,
	}
	SmartReadThresholds = transport.ATACommand{
		Command: 0xB0, Features: 0xD1, LBAMid: 0x4F, LBAHigh: 0xC2, Device: lbaDevice,
	}
)
