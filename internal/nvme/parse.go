package nvme

import (
	"encoding/binary"
	"fmt"
	"strings"

	"vrx/gosmart/model"
)

const (
	CNSController = 0x01
	LogSMART      = 0x02
	NSIDAll       = 0xFFFFFFFF
	HealthLogSize = 512
)

type Controller struct {
	Serial, Model, Firmware string
	TotalCapacity           uint64
}

func ParseIdentifyController(b []byte) (Controller, error) {
	if len(b) < 72 {
		return Controller{}, fmt.Errorf("%w: Identify Controller shorter than 72 bytes (%d)", model.ErrInvalidData, len(b))
	}
	c := Controller{
		Serial:   str(b[4:24]),
		Model:    str(b[24:64]),
		Firmware: str(b[64:72]),
	}
	if len(b) >= 296 {
		c.TotalCapacity = binary.LittleEndian.Uint64(b[280:])
	}
	return c, nil
}

type HealthLog struct {
	Health       model.NVMeHealth
	TemperatureK uint16
	PowerCycles  uint64
	PowerOnHours uint64
}

func ParseHealthLog(b []byte) (HealthLog, error) {
	if len(b) < 192 {
		return HealthLog{}, fmt.Errorf("%w: SMART log shorter than 192 bytes (%d)", model.ErrInvalidData, len(b))
	}
	u64 := func(off int) uint64 { return binary.LittleEndian.Uint64(b[off:]) }
	return HealthLog{
		TemperatureK: binary.LittleEndian.Uint16(b[1:3]),
		PowerCycles:  u64(112),
		PowerOnHours: u64(128),
		Health: model.NVMeHealth{
			CriticalWarning:   b[0],
			AvailableSpare:    b[3],
			SpareThreshold:    b[4],
			PercentageUsed:    b[5],
			DataUnitsRead:     u64(32),
			DataUnitsWritten:  u64(48),
			HostReads:         u64(64),
			HostWrites:        u64(80),
			ControllerBusyMin: u64(96),
			UnsafeShutdowns:   u64(144),
			MediaErrors:       u64(160),
			ErrorLogEntries:   u64(176),
		},
	}, nil
}

func str(b []byte) string { return strings.TrimSpace(strings.Trim(string(b), "\x00")) }
