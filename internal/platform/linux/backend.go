//go:build linux

package linux

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"

	"github.com/rnmz/gosmart/internal/transport"
	"github.com/rnmz/gosmart/model"
)

var reDisk = regexp.MustCompile(`^(sd[a-z]+|hd[a-z]+|vd[a-z]+|nvme\d+n\d+)$`)

type Backend struct {
	sysBlock string
}

func New() *Backend { return &Backend{sysBlock: "/sys/block"} }

func (b *Backend) List() ([]string, error) {
	ents, err := os.ReadDir(b.sysBlock)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if reDisk.MatchString(e.Name()) {
			out = append(out, "/dev/"+e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func (b *Backend) Probe(path string) (model.Protocol, error) {
	if strings.HasPrefix(filepath.Base(path), "nvme") {
		return model.ProtocolNVMe, nil
	}
	return model.ProtocolATA, nil
}

func (b *Backend) OpenATA(path string) (transport.ATADevice, error) {
	f, err := openDev(path)
	if err != nil {
		return nil, err
	}
	return &ataDevice{path: path, f: f}, nil
}

func (b *Backend) OpenNVMe(path string) (transport.NVMeDevice, error) {
	f, err := openDev(path)
	if err != nil {
		return nil, err
	}
	return &nvmeDevice{path: path, f: f}, nil
}

func openDev(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, model.NewDeviceError(path, "open", err)
	}
	return f, nil
}
