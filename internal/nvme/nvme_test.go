package nvme

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

func logBuf(warning uint8, mediaErrors uint64) []byte {
	b := make([]byte, 512)
	b[0] = warning
	binary.LittleEndian.PutUint16(b[1:], 313)
	b[3], b[4], b[5] = 100, 10, 7
	binary.LittleEndian.PutUint64(b[48:], 1000)
	binary.LittleEndian.PutUint64(b[112:], 55)
	binary.LittleEndian.PutUint64(b[128:], 4321)
	binary.LittleEndian.PutUint64(b[160:], mediaErrors)
	return b
}

func identBuf() []byte {
	b := make([]byte, 4096)
	copy(b[4:], "SERIAL01            ")
	copy(b[24:], "FAST NVME 512GB")
	copy(b[64:], "1.2.3")
	return b
}

type fakeDev struct {
	log []byte
	err error
}

func (f fakeDev) Close() error                               { return nil }
func (f fakeDev) Identify(uint8) ([]byte, error)             { return identBuf(), nil }
func (f fakeDev) LogPage(uint8, uint32, int) ([]byte, error) { return f.log, f.err }

var _ transport.NVMeDevice = fakeDev{}

func TestParseHealthLog(t *testing.T) {
	l, err := ParseHealthLog(logBuf(0, 2))
	if err != nil {
		t.Fatal(err)
	}
	if l.PowerOnHours != 4321 || l.PowerCycles != 55 || l.Health.PercentageUsed != 7 ||
		l.Health.DataUnitsWritten != 1000 || l.Health.MediaErrors != 2 {
		t.Fatalf("%+v", l)
	}
	if _, err := ParseHealthLog(make([]byte, 5)); !errors.Is(err, model.ErrInvalidData) {
		t.Fatal(err)
	}
}

func TestCollect(t *testing.T) {
	info, err := Collect(fakeDev{log: logBuf(0, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if info.Model != "FAST NVME 512GB" || info.Serial != "SERIAL01" || info.Health != model.HealthOK ||
		info.TemperatureC == nil || *info.TemperatureC != 40 {
		t.Fatalf("%+v", info)
	}
	info, _ = Collect(fakeDev{log: logBuf(0x04, 0)})
	if info.Health != model.HealthFailing {
		t.Fatal("critical warning must yield failing")
	}
	info, _ = Collect(fakeDev{log: logBuf(0, 3)})
	if info.Health != model.HealthWarning {
		t.Fatal("media errors must yield warning")
	}
	worn := logBuf(0, 0)
	worn[5] = 95
	info, _ = Collect(fakeDev{log: worn})
	if info.Health != model.HealthWarning {
		t.Fatal("high wear must yield warning")
	}
	if info, err = Collect(fakeDev{err: errors.New("boom")}); err == nil || info.Model == "" {
		t.Fatal("expected a partial result and an error")
	}
}
