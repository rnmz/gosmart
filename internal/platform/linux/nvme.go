//go:build linux

package linux

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"vrx/gosmart/model"
)

const nvmeIoctlAdminCmd = 0xC0484E41

const (
	opGetLogPage = 0x02
	opIdentify   = 0x06
)

type nvmeAdminCmd struct {
	Opcode      uint8
	Flags       uint8
	Rsvd1       uint16
	NSID        uint32
	Cdw2, Cdw3  uint32
	Metadata    uint64
	Addr        uint64
	MetadataLen uint32
	DataLen     uint32
	Cdw10       uint32
	Cdw11       uint32
	Cdw12       uint32
	Cdw13       uint32
	Cdw14       uint32
	Cdw15       uint32
	TimeoutMs   uint32
	Result      uint32
}

type nvmeDevice struct {
	path string
	f    *os.File
}

func (d *nvmeDevice) Close() error { return d.f.Close() }

func (d *nvmeDevice) Identify(cns uint8) ([]byte, error) {
	return d.admin("Identify", &nvmeAdminCmd{Opcode: opIdentify, Cdw10: uint32(cns)}, 4096)
}

func (d *nvmeDevice) LogPage(lid uint8, nsid uint32, size int) ([]byte, error) {
	numd := uint32(size/4 - 1)
	cmd := &nvmeAdminCmd{Opcode: opGetLogPage, NSID: nsid, Cdw10: numd<<16 | uint32(lid)}
	return d.admin(fmt.Sprintf("Get Log Page 0x%02X", lid), cmd, size)
}

func (d *nvmeDevice) admin(op string, c *nvmeAdminCmd, size int) ([]byte, error) {
	buf := make([]byte, size)
	c.Addr = uint64(uintptr(unsafe.Pointer(&buf[0])))
	c.DataLen = uint32(size)
	c.TimeoutMs = 5000
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, d.f.Fd(), nvmeIoctlAdminCmd, uintptr(unsafe.Pointer(c)))
	runtime.KeepAlive(buf)
	if errno != 0 {
		return nil, model.NewDeviceError(d.path, "NVMe "+op, errno)
	}
	return buf, nil
}
