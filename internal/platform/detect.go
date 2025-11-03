// Package platform provides platform detection and system utilities.
package platform

import (
	"runtime"

	"github.com/willibrandon/lazynuget/internal/config"
	"github.com/willibrandon/lazynuget/internal/logging"
)

// Platform provides platform-specific information and utilities.
type Platform interface {
	// OS returns the operating system name
	OS() string

	// Arch returns the architecture
	Arch() string
}

// platformImpl implements the Platform interface.
type platformImpl struct {
	os   string
	arch string
}

func (p *platformImpl) OS() string {
	return p.os
}

func (p *platformImpl) Arch() string {
	return p.arch
}

// New creates a new Platform instance.
// For now, this is a stub that returns basic runtime information.
func New(cfg *config.AppConfig, log logging.Logger) Platform {
	return &platformImpl{
		os:   runtime.GOOS,
		arch: runtime.GOARCH,
	}
}
