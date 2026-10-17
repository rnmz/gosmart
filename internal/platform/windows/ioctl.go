//go:build windows

package windows

import (
	"syscall"
	"unsafe"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procDeviceIoControl = kernel32.NewProc("DeviceIoControl")
)

func deviceIoControl(h syscall.Handle, code uint32, in unsafe.Pointer, inLen uint32, out unsafe.Pointer, outLen uint32) (uint32, error) {
	var returned uint32
	r, _, err := procDeviceIoControl.Call(
		uintptr(h), uintptr(code),
		uintptr(in), uintptr(inLen),
		uintptr(out), uintptr(outLen),
		uintptr(unsafe.Pointer(&returned)), 0,
	)
	if r == 0 {
		return 0, err
	}
	return returned, nil
}

func openPhysical(path string, access uint32) (syscall.Handle, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	return syscall.CreateFile(p, access,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, 0, 0)
}
