package svchealthcheck

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrCheckerPanic = errors.New("checker panicked")
)

type Healthcheck struct {
	checkerTimeout time.Duration
	healthCheckers map[string]Checker
	readyCheckers  map[string]Checker
}

func NewHealthcheck(opts ...Option) *Healthcheck {
	o := defaultOpts()
	for _, opt := range opts {
		opt(&o)
	}
	r := &Healthcheck{
		checkerTimeout: o.timeout,
		healthCheckers: o.healthCheckers,
		readyCheckers:  o.readyCheckers,
	}
	return r
}

func (s *Healthcheck) Health(ctx context.Context) *CheckResponse {
	return s.generateResponse(ctx, s.healthCheckers)
}

func (s *Healthcheck) Ready(ctx context.Context) *CheckResponse {
	return s.generateResponse(ctx, s.readyCheckers)
}

func (s *Healthcheck) generateResponse(ctx context.Context, checks map[string]Checker) *CheckResponse {
	if s.checkerTimeout > 0 {
		ctx2, cancel := context.WithTimeout(ctx, s.checkerTimeout)
		defer cancel()
		ctx = ctx2
	}

	jsonResponse := &CheckResponse{
		StatusCode: http.StatusOK,
		Checks:     make(map[string]CheckResponseEntry, len(checks)),
	}

	var (
		wg      sync.WaitGroup
		checksM sync.Mutex
	)
	wg.Add(len(checks))
	for key, check := range checks {
		go func(key string, check Checker) {
			defer wg.Done()
			st := time.Now()
			errch := make(chan error, 1)

			// Start another goroutine to be able to track timeouts.
			go func() {
				defer func() {
					r := recover()
					if r == nil {
						return
					}
					handlerRecover(r, errch)
				}()

				errch <- check.Check(ctx)
			}()

			var err error
			select {
			case <-ctx.Done(): // timeout
				err = ctx.Err()
			case e := <-errch:
				err = e
			}

			checksM.Lock()
			if err != nil { // If check fails, return service unavailable.
				jsonResponse.StatusCode = errorToStatus(jsonResponse.StatusCode, err)
			}
			jsonResponse.Checks[key] = CheckResponseEntry{
				Duration: time.Since(st),
				Error:    errorMessage(err),
			}
			checksM.Unlock()
		}(key, check)
	}

	wg.Wait() // Wait for all checks to finish.

	jsonResponse.Status = http.StatusText(jsonResponse.StatusCode)

	return jsonResponse
}

func errorToStatus(code int, err error) int {
	if strings.HasPrefix(err.Error(), ErrCheckerPanic.Error()) {
		return http.StatusInternalServerError
	}
	if code == http.StatusOK {
		return http.StatusServiceUnavailable
	}
	return code
}

// handlerRecover handles a possible panic from the handler implementation.
func handlerRecover(r interface{}, errch chan<- error) {
	if r == nil {
		return
	}

	errch <- panicError(r)
}

// errorMessage returns the error message for the given error. If the error is nil, the message returned is empty.
func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// panicError returns the panic message for the given panic. If the recover data is an error, the method
// errorMessage is used.
func panicError(r interface{}) error {
	err, ok := r.(error)
	if ok {
		r = err.Error()
	}

	return fmt.Errorf("%s: %v", ErrCheckerPanic, r)
}
