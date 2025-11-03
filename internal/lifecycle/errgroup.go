package lifecycle

import (
	"context"
	"runtime/debug"

	"github.com/willibrandon/lazynuget/internal/logging"
	"golang.org/x/sync/errgroup"
)

// ErrorGroup wraps errgroup.Group with panic recovery (Layer 4)
type ErrorGroup struct {
	group  *errgroup.Group
	ctx    context.Context
	logger logging.Logger
}

// NewErrorGroup creates a new ErrorGroup with context
func NewErrorGroup(ctx context.Context, logger logging.Logger) *ErrorGroup {
	group, groupCtx := errgroup.WithContext(ctx)
	return &ErrorGroup{
		group:  group,
		ctx:    groupCtx,
		logger: logger,
	}
}

// Go launches a goroutine with panic recovery
func (eg *ErrorGroup) Go(name string, fn func(context.Context) error) {
	eg.group.Go(func() error {
		// Layer 4 panic recovery: Protect goroutines
		defer func() {
			if r := recover(); r != nil {
				if eg.logger != nil {
					eg.logger.Error("PANIC in goroutine '%s': %v\nStack: %s", name, r, debug.Stack())
				}
			}
		}()

		if eg.logger != nil {
			eg.logger.Debug("Starting goroutine: %s", name)
		}

		err := fn(eg.ctx)

		if err != nil && eg.logger != nil {
			eg.logger.Error("Goroutine '%s' failed: %v", name, err)
		}

		return err
	})
}

// Wait blocks until all goroutines have completed
func (eg *ErrorGroup) Wait() error {
	return eg.group.Wait()
}

// Context returns the context associated with the error group
func (eg *ErrorGroup) Context() context.Context {
	return eg.ctx
}
