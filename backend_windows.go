//go:build windows

package gosmart

import (
	"github.com/rnmz/gosmart/internal/platform/windows"
	"github.com/rnmz/gosmart/internal/transport"
)

func defaultBackend() transport.Backend { return windows.New() }
