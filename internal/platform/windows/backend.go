//go:build windows

package windows

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"syscall"
	"unsafe"

	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

const (
	maxDrives = 64

	ioctlStorageQueryProperty = 0x002D1400
	storageDeviceProperty     = 0
	propertyStandardQuery     = 0
	busTypeNvme               = 17
)

type Backend struct{}

func New() *Backend { return &Backend{} }

func drivePath(i int) string { return fmt.Sprintf(`\\.\PhysicalDrive%d`, i) }

func (b *Backend) List() ([]string, error) {
	var out []string
	for i := 0; i < maxDrives; i++ {
		p := drivePath(i)
		h, err := openPhysical(p, 0)
		if err == nil {
			syscall.CloseHandle(h)
			out = append(out, p)
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (b *Backend) Probe(path string) (model.Protocol, error) {
	h, err := openPhysical(path, 0)
	if err != nil {
		return "", model.NewDeviceError(path, "open", err)
	}
	defer syscall.CloseHandle(h)

	query := make([]byte, 12)
	binary.LittleEndian.PutUint32(query[0:], storageDeviceProperty)
	binary.LittleEndian.PutUint32(query[4:], propertyStandardQuery)
	out := make([]byte, 1024)
	if _, err := deviceIoControl(h, ioctlStorageQueryProperty,
		unsafe.Pointer(&query[0]), uint32(len(query)),
		unsafe.Pointer(&out[0]), uint32(len(out))); err != nil {
		return "", model.NewDeviceError(path, "query device property", err)
	}
	if binary.LittleEndian.Uint32(out[28:]) == busTypeNvme {
		return model.ProtocolNVMe, nil
	}
	return model.ProtocolATA, nil
}

func (b *Backend) OpenATA(path string) (transport.ATADevice, error) {
	h, err := openRW(path)
	if err != nil {
		return nil, err
	}
	return &ataDevice{path: path, h: h}, nil
}

func (b *Backend) OpenNVMe(path string) (transport.NVMeDevice, error) {
	h, err := openRW(path)
	if err != nil {
		return nil, err
	}
	return &nvmeDevice{path: path, h: h}, nil
}

func openRW(path string) (syscall.Handle, error) {
	h, err := openPhysical(path, syscall.GENERIC_READ|syscall.GENERIC_WRITE)
	if err != nil {
		return syscall.InvalidHandle, model.NewDeviceError(path, "open", err)
	}
	return h, nil
}
