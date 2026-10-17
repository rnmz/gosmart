package ata

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/rnmz/gosmart/model"
)

const sectorLen = 512

type Identity struct {
	Model, Serial, Firmware string
	Sectors                 uint64
	SectorSize              uint32
	RotationRate            uint16
	SMARTSupported          bool
	SMARTEnabled            bool
}

func (i Identity) CapacityBytes() uint64 { return i.Sectors * uint64(i.SectorSize) }

func ParseIdentify(b []byte) (Identity, error) {
	if len(b) < sectorLen {
		return Identity{}, fmt.Errorf("%w: IDENTIFY response shorter than 512 bytes (%d)", model.ErrInvalidData, len(b))
	}
	w := func(i int) uint16 { return binary.LittleEndian.Uint16(b[i*2:]) }

	id := Identity{
		Serial:         decodeString(b[10*2 : 20*2]),
		Firmware:       decodeString(b[23*2 : 27*2]),
		Model:          decodeString(b[27*2 : 47*2]),
		RotationRate:   w(217),
		SMARTSupported: w(82)&1 != 0,
		SMARTEnabled:   w(85)&1 != 0,
		SectorSize:     512,
	}
	id.Sectors = uint64(w(100)) | uint64(w(101))<<16 | uint64(w(102))<<32 | uint64(w(103))<<48
	if id.Sectors == 0 {
		id.Sectors = uint64(w(60)) | uint64(w(61))<<16
	}

	if w(106)&0xC000 == 0x4000 && w(106)&(1<<12) != 0 {
		if v := uint32(w(117)) | uint32(w(118))<<16; v != 0 {
			id.SectorSize = v * 2
		}
	}
	return id, nil
}

func decodeString(b []byte) string {
	out := make([]byte, 0, len(b))
	for i := 0; i+1 < len(b); i += 2 {
		out = append(out, b[i+1], b[i])
	}
	return strings.TrimSpace(strings.Trim(string(out), "\x00"))
}
