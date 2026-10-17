package ata

import (
	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

func Collect(dev transport.ATADevice) (*model.Info, error) {
	raw, err := dev.ReadATA(IdentifyDevice)
	if err != nil {
		return nil, err
	}
	id, err := ParseIdentify(raw)
	if err != nil {
		return nil, err
	}
	info := &model.Info{
		Protocol:       model.ProtocolATA,
		Model:          id.Model,
		Serial:         id.Serial,
		Firmware:       id.Firmware,
		CapacityBytes:  id.CapacityBytes(),
		RotationRate:   id.RotationRate,
		SMARTSupported: id.SMARTSupported,
		Health:         model.HealthUnknown,
	}
	if !id.SMARTSupported {
		return info, nil
	}

	data, err := dev.ReadATA(SmartReadData)
	if err != nil {
		return info, err
	}
	thresholds, _ := dev.ReadATA(SmartReadThresholds)
	attrs, err := ParseAttributes(data, thresholds)
	if err != nil {
		return info, err
	}
	s := Summarize(attrs)
	info.Attributes = attrs
	info.Health = s.Health
	info.TemperatureC = s.TemperatureC
	info.PowerOnHours = s.PowerOnHours
	info.PowerCycles = s.PowerCycles
	return info, nil
}
