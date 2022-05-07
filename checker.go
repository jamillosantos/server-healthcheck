package svchealthcheck

import "context"

// Checker checks the health of a dependency and return an error if it is not healthy.
type Checker interface {
	Check(ctx context.Context) error
}

// CheckerFunc is a Checker defined as a function.
type CheckerFunc func(ctx context.Context) error

// Check implements the Checker interface for the CheckerFunc type.
func (c CheckerFunc) Check(ctx context.Context) error {
	return c(ctx)
}
