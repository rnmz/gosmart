//go:build linux

package linux

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

const (
	sgIO             = 0x2285
	sgDxferFromDev   = -3
	ataPassThrough16 = 0x85
	sectorLen        = 512
)

type sgIOHdr struct {
	InterfaceID    int32
	DxferDirection int32
	CmdLen         uint8
	MxSbLen        uint8
	IovecCount     uint16
	DxferLen       uint32
	Dxferp         unsafe.Pointer
	Cmdp           *byte
	Sbp            *byte
	Timeout        uint32
	Flags          uint32
	PackID         int32
	UsrPtr         unsafe.Pointer
	Status         uint8
	MaskedStatus   uint8
	MsgStatus      uint8
	SbLenWr        uint8
	HostStatus     uint16
	DriverStatus   uint16
	Resid          int32
	Duration       uint32
	Info           uint32
}

type ataDevice struct {
	path string
	f    *os.File
}

func (d *ataDevice) Close() error { return d.f.Close() }

func (d *ataDevice) ReadATA(c transport.ATACommand) ([]byte, error) {
	op := fmt.Sprintf("ATA 0x%02X", c.Command)
	buf := make([]byte, sectorLen)
	cdb := [16]byte{
		ataPassThrough16,
		4 << 1,
		0x0e,
		0, c.Features,
		0, 1,
		0, c.LBALow,
		0, c.LBAMid,
		0, c.LBAHigh,
		c.Device,
		c.Command,
		0,
	}
	var sense [32]byte
	h := sgIOHdr{
		InterfaceID:    'S',
		DxferDirection: sgDxferFromDev,
		CmdLen:         uint8(len(cdb)),
		MxSbLen:        uint8(len(sense)),
		DxferLen:       uint32(len(buf)),
		Dxferp:         unsafe.Pointer(&buf[0]),
		Cmdp:           &cdb[0],
		Sbp:            &sense[0],
		Timeout:        5000,
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, d.f.Fd(), sgIO, uintptr(unsafe.Pointer(&h)))
	if errno != 0 {
		return nil, model.NewDeviceError(d.path, op, errno)
	}
	if h.Status != 0 || h.HostStatus != 0 {
		return nil, model.NewDeviceError(d.path, op,
			fmt.Errorf("status=%#x host=%#x driver=%#x", h.Status, h.HostStatus, h.DriverStatus))
	}
	return buf, nil
}
