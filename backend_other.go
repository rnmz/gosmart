//go:build !linux && !windows

package gosmart

import (
	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

type unsupported struct{}

func defaultBackend() transport.Backend { return unsupported{} }

func (unsupported) List() ([]string, error) { return nil, model.ErrUnsupportedPlatform }
func (unsupported) Probe(string) (model.Protocol, error) {
	return "", model.ErrUnsupportedPlatform
}
func (unsupported) OpenATA(string) (transport.ATADevice, error) {
	return nil, model.ErrUnsupportedPlatform
}
func (unsupported) OpenNVMe(string) (transport.NVMeDevice, error) {
	return nil, model.ErrUnsupportedPlatform
}
