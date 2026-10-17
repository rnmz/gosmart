package ata

import (
	"encoding/binary"
	"errors"
	"testing"

	"vrx/gosmart/internal/transport"
	"vrx/gosmart/model"
)

func putString(b []byte, s string) {
	for i := range b {
		b[i] = ' '
	}
	copy(b, s)
	for i := 0; i+1 < len(b); i += 2 {
		b[i], b[i+1] = b[i+1], b[i]
	}
}

func identifyBuf() []byte {
	b := make([]byte, 512)
	putString(b[10*2:20*2], "SN12345")
	putString(b[23*2:27*2], "FW01")
	putString(b[27*2:47*2], "TEST DISK 1TB")
	binary.LittleEndian.PutUint16(b[82*2:], 1)
	binary.LittleEndian.PutUint16(b[85*2:], 1)
	binary.LittleEndian.PutUint16(b[100*2:], 2000)
	return b
}

func entry(tbl []byte, idx int, id uint8, flags uint16, cur, worst uint8, raw uint64) {
	e := tbl[2+idx*12:]
	e[0] = id
	binary.LittleEndian.PutUint16(e[1:], flags)
	e[3], e[4] = cur, worst
	for j := 0; j < 6; j++ {
		e[5+j] = byte(raw >> (8 * j))
	}
}

func smartTables(reallocCur uint8, reallocRaw uint64) (data, thr []byte) {
	data, thr = make([]byte, 512), make([]byte, 512)
	entry(data, 0, 5, 0x0033, reallocCur, 100, reallocRaw)
	entry(data, 1, 9, 0x0032, 99, 99, 12345)
	entry(data, 2, 12, 0x0032, 99, 99, 77)
	entry(data, 3, 194, 0x0022, 60, 50, 40)
	thr[2], thr[3] = 5, 36
	return
}

func TestParseIdentify(t *testing.T) {
	id, err := ParseIdentify(identifyBuf())
	if err != nil {
		t.Fatal(err)
	}
	if id.Model != "TEST DISK 1TB" || id.Serial != "SN12345" || id.Firmware != "FW01" {
		t.Fatalf("strings: %+v", id)
	}
	if id.CapacityBytes() != 2000*512 || !id.SMARTSupported || !id.SMARTEnabled {
		t.Fatalf("%+v", id)
	}
	if _, err := ParseIdentify(make([]byte, 10)); !errors.Is(err, model.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestParseAttributesAndSummary(t *testing.T) {
	data, thr := smartTables(100, 0)
	attrs, err := ParseAttributes(data, thr)
	if err != nil || len(attrs) != 4 {
		t.Fatalf("%v %d", err, len(attrs))
	}
	s := Summarize(attrs)
	if s.Health != model.HealthOK || s.PowerOnHours != 12345 || s.PowerCycles != 77 ||
		s.TemperatureC == nil || *s.TemperatureC != 40 {
		t.Fatalf("%+v", s)
	}

	data, thr = smartTables(100, 12)
	attrs, _ = ParseAttributes(data, thr)
	if Summarize(attrs).Health != model.HealthWarning {
		t.Fatal("expected warning")
	}

	data, thr = smartTables(30, 12)
	attrs, _ = ParseAttributes(data, thr)
	if Summarize(attrs).Health != model.HealthFailing {
		t.Fatal("expected failing")
	}
}

type fakeDev struct{ resp map[uint8][]byte }

func (f fakeDev) Close() error { return nil }
func (f fakeDev) ReadATA(c transport.ATACommand) ([]byte, error) {
	key := c.Command
	if c.Command == 0xB0 {
		key = c.Features
	}
	b, ok := f.resp[key]
	if !ok {
		return nil, errors.New("no response")
	}
	return b, nil
}

func TestCollect(t *testing.T) {
	data, thr := smartTables(100, 0)
	info, err := Collect(fakeDev{resp: map[uint8][]byte{0xEC: identifyBuf(), 0xD0: data, 0xD1: thr}})
	if err != nil {
		t.Fatal(err)
	}
	if info.Model != "TEST DISK 1TB" || info.Health != model.HealthOK || len(info.Attributes) != 4 {
		t.Fatalf("%+v", info)
	}

	info, err = Collect(fakeDev{resp: map[uint8][]byte{0xEC: identifyBuf()}})
	if err == nil || info == nil || info.Model == "" || info.Health != model.HealthUnknown {
		t.Fatalf("partial result: %+v err=%v", info, err)
	}
}
