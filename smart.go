package gosmart

import (
	"context"
	"fmt"

	"github.com/rnmz/gosmart/internal/ata"
	"github.com/rnmz/gosmart/internal/nvme"
	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

type (
	Info        = model.Info
	Attribute   = model.Attribute
	NVMeHealth  = model.NVMeHealth
	Protocol    = model.Protocol
	Health      = model.Health
	DeviceError = model.DeviceError
)

const (
	ProtocolATA   = model.ProtocolATA
	ProtocolNVMe  = model.ProtocolNVMe
	HealthUnknown = model.HealthUnknown
	HealthOK      = model.HealthOK
	HealthWarning = model.HealthWarning
	HealthFailing = model.HealthFailing
)

var (
	ErrPermission          = model.ErrPermission
	ErrUnsupportedPlatform = model.ErrUnsupportedPlatform
	ErrInvalidData         = model.ErrInvalidData
)

type Result struct {
	Path string
	Info *Info
	Err  error
}

type Reader struct {
	backend transport.Backend
}

func New() *Reader { return newWithBackend(defaultBackend()) }

func newWithBackend(b transport.Backend) *Reader { return &Reader{backend: b} }

func (r *Reader) List() ([]string, error) { return r.backend.List() }

func (r *Reader) Read(path string) (*Info, error) {
	proto, err := r.backend.Probe(path)
	if err != nil {
		return nil, err
	}
	var info *Info
	switch proto {
	case model.ProtocolATA:
		dev, oerr := r.backend.OpenATA(path)
		if oerr != nil {
			return nil, oerr
		}
		defer dev.Close()
		info, err = ata.Collect(dev)
	case model.ProtocolNVMe:
		dev, oerr := r.backend.OpenNVMe(path)
		if oerr != nil {
			return nil, oerr
		}
		defer dev.Close()
		info, err = nvme.Collect(dev)
	default:
		return nil, fmt.Errorf("%s: unknown protocol %q", path, proto)
	}
	if info != nil {
		info.Path = path
	}
	return info, err
}

func (r *Reader) ReadAll(ctx context.Context) ([]Result, error) {
	paths, err := r.List()
	if err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		info, rerr := r.Read(p)
		results = append(results, Result{Path: p, Info: info, Err: rerr})
	}
	return results, nil
}
