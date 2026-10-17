package nvme

import (
	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

func Collect(dev transport.NVMeDevice) (*model.Info, error) {
	info := &model.Info{
		Protocol:       model.ProtocolNVMe,
		SMARTSupported: true,
		Health:         model.HealthUnknown,
	}
	if raw, err := dev.Identify(CNSController); err == nil {
		if c, err := ParseIdentifyController(raw); err == nil {
			info.Model, info.Serial, info.Firmware = c.Model, c.Serial, c.Firmware
			info.CapacityBytes = c.TotalCapacity
		}
	}

	raw, err := dev.LogPage(LogSMART, NSIDAll, HealthLogSize)
	if err != nil {
		return info, err
	}
	log, err := ParseHealthLog(raw)
	if err != nil {
		return info, err
	}
	h := log.Health
	info.NVMe = &h
	info.PowerCycles = log.PowerCycles
	info.PowerOnHours = log.PowerOnHours
	if log.TemperatureK > 0 {
		t := int(log.TemperatureK) - 273
		info.TemperatureC = &t
	}
	info.Health = model.HealthOK
	switch {
	case h.CriticalWarning != 0:
		info.Health = model.HealthFailing
	case h.MediaErrors > 0 || h.PercentageUsed >= 90:
		info.Health = model.HealthWarning
	}
	return info, nil
}
