package model

import (
	"errors"
	"fmt"
	"io/fs"
)

var (
	ErrPermission          = errors.New("insufficient privileges (run as root / Administrator)")
	ErrUnsupportedPlatform = errors.New("platform not supported")
	ErrInvalidData         = errors.New("invalid device response")
)

type DeviceError struct {
	Path string
	Op   string
	Err  error
}

func (e *DeviceError) Error() string { return fmt.Sprintf("%s: %s: %v", e.Path, e.Op, e.Err) }
func (e *DeviceError) Unwrap() error { return e.Err }

func NewDeviceError(path, op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, fs.ErrPermission) && !errors.Is(err, ErrPermission) {
		err = fmt.Errorf("%w: %w", ErrPermission, err)
	}
	return &DeviceError{Path: path, Op: op, Err: err}
}
