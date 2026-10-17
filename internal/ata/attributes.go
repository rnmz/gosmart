package ata

import (
	"encoding/binary"
	"fmt"

	"github.com/rnmz/gosmart/model"
)

const (
	tableOffset = 2
	entrySize   = 12
	entryCount  = 30
)

func ParseAttributes(data, thresholds []byte) ([]model.Attribute, error) {
	if len(data) < sectorLen {
		return nil, fmt.Errorf("%w: SMART data shorter than 512 bytes (%d)", model.ErrInvalidData, len(data))
	}
	thr := map[uint8]uint8{}
	if len(thresholds) >= sectorLen {
		for i := 0; i < entryCount; i++ {
			e := thresholds[tableOffset+i*entrySize:]
			if e[0] != 0 {
				thr[e[0]] = e[1]
			}
		}
	}
	var attrs []model.Attribute
	for i := 0; i < entryCount; i++ {
		e := data[tableOffset+i*entrySize : tableOffset+(i+1)*entrySize]
		if e[0] == 0 {
			continue
		}
		flags := binary.LittleEndian.Uint16(e[1:3])
		a := model.Attribute{
			ID:        e[0],
			Name:      AttributeName(e[0]),
			Flags:     flags,
			PreFail:   flags&1 != 0,
			Current:   e[3],
			Worst:     e[4],
			Threshold: thr[e[0]],
		}
		for j := 0; j < 6; j++ {
			a.Raw |= uint64(e[5+j]) << (8 * j)
		}
		attrs = append(attrs, a)
	}
	return attrs, nil
}

type Summary struct {
	Health       model.Health
	TemperatureC *int
	PowerOnHours uint64
	PowerCycles  uint64
}

var warnAttributes = map[uint8]bool{5: true, 187: true, 197: true, 198: true}

func Summarize(attrs []model.Attribute) Summary {
	s := Summary{Health: model.HealthUnknown}
	if len(attrs) > 0 {
		s.Health = model.HealthOK
	}
	var t190, t194 *int
	warn := false
	for _, a := range attrs {
		if a.Failing() {
			s.Health = model.HealthFailing
		}
		if warnAttributes[a.ID] && a.Raw > 0 {
			warn = true
		}
		switch a.ID {
		case 9:
			s.PowerOnHours = a.Raw & 0xFFFFFFFF
		case 12:
			s.PowerCycles = a.Raw & 0xFFFFFFFF
		case 190:
			t190 = tempFromRaw(a.Raw)
		case 194:
			t194 = tempFromRaw(a.Raw)
		}
	}
	if warn && s.Health == model.HealthOK {
		s.Health = model.HealthWarning
	}
	if s.TemperatureC = t194; s.TemperatureC == nil {
		s.TemperatureC = t190
	}
	return s
}

func tempFromRaw(raw uint64) *int {
	t := int(raw & 0xFF)
	if t == 0 || t >= 150 {
		return nil
	}
	return &t
}
