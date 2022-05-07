package svchealthcheck

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestWithBindAddress(t *testing.T) {
	var opts options
	wantBindAddress := "bind address"
	WithBindAddress(wantBindAddress)(&opts)
	assert.Equal(t, wantBindAddress, opts.bindAddress)
}

func TestWithInitializer(t *testing.T) {
	var opts options
	wantInitializer := func(app *fiber.App) error {
		return nil
	}
	WithInitializer(wantInitializer)(&opts)
	assert.Equal(t, fmt.Sprintf("%p", wantInitializer), fmt.Sprintf("%p", opts.initializer))
}

func TestWithTimeout(t *testing.T) {
	var opts options
	wantTimeout := time.Second
	WithTimeout(wantTimeout)(&opts)
	assert.Equal(t, wantTimeout, opts.timeout)
}

func TestWithCheck(t *testing.T) {
	opts := defaultOpts()
	wantCheck := CheckerFunc(func(ctx context.Context) error { return nil })
	WithCheck("check1", wantCheck)(&opts)
	assert.Contains(t, opts.healthCheckers, "check1")
}

func TestWithReadyCheck(t *testing.T) {
	opts := defaultOpts()
	wantCheck := CheckerFunc(func(ctx context.Context) error { return nil })
	WithReadyCheck("check1", wantCheck)(&opts)
	assert.Contains(t, opts.readyCheckers, "check1")
}

func Test_options_GetBindAddress(t *testing.T) {
	wantBindAddr := "bind address"
	opts := options{bindAddress: wantBindAddr}
	gotBindAddr := opts.GetBindAddress()
	assert.Equal(t, wantBindAddr, gotBindAddr)
}
