package platform

import "sync"

var (
	// Singleton instance
	instance PlatformInfo
	once     sync.Once
	initErr  error
)

// New returns the singleton PlatformInfo instance
// Platform detection is performed once and cached
func New() (PlatformInfo, error) {
	once.Do(func() {
		instance, initErr = detectPlatform()
	})

	return instance, initErr
}
