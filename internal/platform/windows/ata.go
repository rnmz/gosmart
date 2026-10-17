//go:build windows

package windows

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

const (
	ioctlATAPassThrough  = 0x0004D02C
	ataFlagsDRDYRequired = 0x01
	ataFlagsDataIn       = 0x02
	sectorLen            = 512
)

type ataPassThroughEx struct {
	Length             uint16
	AtaFlags           uint16
	PathID             uint8
	TargetID           uint8
	Lun                uint8
	ReservedAsUchar    uint8
	DataTransferLength uint32
	TimeOutValue       uint32
	ReservedAsUlong    uint32
	DataBufferOffset   uintptr
	PreviousTaskFile   [8]byte
	CurrentTaskFile    [8]byte
}

type ataPassThroughBuf struct {
	hdr  ataPassThroughEx
	data [sectorLen]byte
}

type ataDevice struct {
	path string
	h    syscall.Handle
}

func (d *ataDevice) Close() error { return syscall.CloseHandle(d.h) }

func (d *ataDevice) ReadATA(c transport.ATACommand) ([]byte, error) {
	op := fmt.Sprintf("ATA 0x%02X", c.Command)
	var buf ataPassThroughBuf
	buf.hdr = ataPassThroughEx{
		Length:             uint16(unsafe.Sizeof(buf.hdr)),
		AtaFlags:           ataFlagsDataIn | ataFlagsDRDYRequired,
		DataTransferLength: sectorLen,
		TimeOutValue:       5,
		DataBufferOffset:   unsafe.Offsetof(buf.data),

		CurrentTaskFile: [8]byte{c.Features, 1, c.LBALow, c.LBAMid, c.LBAHigh, c.Device, c.Command, 0},
	}
	size := uint32(unsafe.Sizeof(buf))
	if _, err := deviceIoControl(d.h, ioctlATAPassThrough,
		unsafe.Pointer(&buf), size, unsafe.Pointer(&buf), size); err != nil {
		return nil, model.NewDeviceError(d.path, op, err)
	}
	if st := buf.hdr.CurrentTaskFile[6]; st&0x01 != 0 {
		return nil, model.NewDeviceError(d.path, op, fmt.Errorf("device returned an error (status=%#x)", st))
	}
	out := make([]byte, sectorLen)
	copy(out, buf.data[:])
	return out, nil
}
