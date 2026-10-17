//go:build linux

package gosmart

import (
	"vrx/gosmart/internal/platform/linux"
	"vrx/gosmart/internal/transport"
)

func defaultBackend() transport.Backend { return linux.New() }
