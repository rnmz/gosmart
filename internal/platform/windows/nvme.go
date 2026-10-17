//go:build windows

package windows

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/rnmz/gosmart/model"
)

const (
	storageAdapterProtocolSpecificProperty = 49
	storageDeviceProtocolSpecificProperty  = 50

	protocolTypeNvme  = 3
	nvmeDataTypeIdent = 1
	nvmeDataTypeLogPg = 2

	queryHeaderLen  = 8
	protocolDataLen = 40
	protocolDataOff = queryHeaderLen + protocolDataLen
)

type nvmeDevice struct {
	path string
	h    syscall.Handle
}

func (d *nvmeDevice) Close() error { return syscall.CloseHandle(d.h) }

func (d *nvmeDevice) Identify(cns uint8) ([]byte, error) {
	return d.queryAny("NVMe Identify",
		[]uint32{storageAdapterProtocolSpecificProperty, storageDeviceProtocolSpecificProperty},
		nvmeDataTypeIdent, uint32(cns), 4096)
}

func (d *nvmeDevice) LogPage(lid uint8, _ uint32, size int) ([]byte, error) {
	return d.queryAny(fmt.Sprintf("NVMe Get Log Page 0x%02X", lid),
		[]uint32{storageDeviceProtocolSpecificProperty, storageAdapterProtocolSpecificProperty},
		nvmeDataTypeLogPg, uint32(lid), size)
}

func (d *nvmeDevice) queryAny(op string, props []uint32, dataType, reqValue uint32, size int) ([]byte, error) {
	var last error
	for _, p := range props {
		out, err := d.query(p, dataType, reqValue, size)
		if err == nil {
			return out, nil
		}
		last = err
	}
	return nil, model.NewDeviceError(d.path, op, last)
}

func (d *nvmeDevice) query(propID, dataType, reqValue uint32, size int) ([]byte, error) {
	buf := make([]byte, protocolDataOff+size)
	le := binary.LittleEndian
	le.PutUint32(buf[0:], propID)
	le.PutUint32(buf[4:], propertyStandardQuery)
	ps := buf[queryHeaderLen:]
	le.PutUint32(ps[0:], protocolTypeNvme)
	le.PutUint32(ps[4:], dataType)
	le.PutUint32(ps[8:], reqValue)
	le.PutUint32(ps[12:], 0)
	le.PutUint32(ps[16:], protocolDataLen)
	le.PutUint32(ps[20:], uint32(size))

	n := uint32(len(buf))
	if _, err := deviceIoControl(d.h, ioctlStorageQueryProperty,
		unsafe.Pointer(&buf[0]), n, unsafe.Pointer(&buf[0]), n); err != nil {
		return nil, err
	}

	off := queryHeaderLen + int(le.Uint32(buf[queryHeaderLen+16:]))
	length := int(le.Uint32(buf[queryHeaderLen+20:]))
	if length > size {
		length = size
	}
	if off < queryHeaderLen || length <= 0 || off+length > len(buf) {
		return nil, fmt.Errorf("%w: invalid offset/length in response (%d/%d)", model.ErrInvalidData, off, length)
	}
	return append([]byte(nil), buf[off:off+length]...), nil
}
