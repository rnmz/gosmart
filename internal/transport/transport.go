package transport

import "github.com/rnmz/gosmart/model"

type ATACommand struct {
	Command  uint8
	Features uint8
	LBALow   uint8
	LBAMid   uint8
	LBAHigh  uint8
	Device   uint8
}

type ATADevice interface {
	ReadATA(cmd ATACommand) ([]byte, error)
	Close() error
}

type NVMeDevice interface {
	Identify(cns uint8) ([]byte, error)

	LogPage(lid uint8, nsid uint32, size int) ([]byte, error)
	Close() error
}

type Backend interface {
	List() ([]string, error)
	Probe(path string) (model.Protocol, error)
	OpenATA(path string) (ATADevice, error)
	OpenNVMe(path string) (NVMeDevice, error)
}
