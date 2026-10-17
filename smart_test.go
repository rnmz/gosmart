package gosmart

import (
	"context"
	"encoding/binary"
	"errors"
	"io/fs"
	"testing"

	"vrx/gosmart/internal/transport"
	"vrx/gosmart/model"
)

type fakeBackend struct{}

func (fakeBackend) List() ([]string, error) {
	return []string{"/dev/nvme0n1", "/dev/sdz", "/dev/denied"}, nil
}
func (fakeBackend) Probe(p string) (model.Protocol, error) {
	if p == "/dev/nvme0n1" {
		return model.ProtocolNVMe, nil
	}
	return model.ProtocolATA, nil
}
func (fakeBackend) OpenATA(p string) (transport.ATADevice, error) {
	if p == "/dev/denied" {
		return nil, model.NewDeviceError(p, "open", fs.ErrPermission)
	}
	return ataFake{}, nil
}
func (fakeBackend) OpenNVMe(string) (transport.NVMeDevice, error) { return nvmeFake{}, nil }

type ataFake struct{}

func (ataFake) Close() error { return nil }
func (ataFake) ReadATA(c transport.ATACommand) ([]byte, error) {
	b := make([]byte, 512)
	if c.Command == 0xEC {
		copy(b[27*2:], "AB")
	}
	return b, nil
}

type nvmeFake struct{}

func (nvmeFake) Close() error                   { return nil }
func (nvmeFake) Identify(uint8) ([]byte, error) { return make([]byte, 4096), nil }
func (nvmeFake) LogPage(uint8, uint32, int) ([]byte, error) {
	b := make([]byte, 512)
	binary.LittleEndian.PutUint16(b[1:], 300)
	return b, nil
}

func TestReadAll(t *testing.T) {
	r := newWithBackend(fakeBackend{})
	res, err := r.ReadAll(context.Background())
	if err != nil || len(res) != 3 {
		t.Fatalf("%v %d", err, len(res))
	}
	if res[0].Info == nil || res[0].Info.Protocol != ProtocolNVMe || res[0].Info.Path != "/dev/nvme0n1" {
		t.Fatalf("%+v", res[0])
	}
	if res[1].Err != nil || res[1].Info.SMARTSupported {
		t.Fatalf("%+v", res[1])
	}
	if !errors.Is(res[2].Err, ErrPermission) {
		t.Fatalf("expected ErrPermission, got %v", res[2].Err)
	}
}

func TestReadAllCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := newWithBackend(fakeBackend{}).ReadAll(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
