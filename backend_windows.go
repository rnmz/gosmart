//go:build windows

package gosmart

import (
	"vrx/gosmart/internal/platform/windows"
	"vrx/gosmart/internal/transport"
)

func defaultBackend() transport.Backend { return windows.New() }
