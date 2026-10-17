//go:build linux

package gosmart

import (
	"github.com/rnmz/gosmart/internal/platform/linux"
	"github.com/rnmz/gosmart/internal/transport"
)

func defaultBackend() transport.Backend { return linux.New() }
