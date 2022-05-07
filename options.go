package svchealthcheck

import (
	"time"

	srvfiber "github.com/jamillosantos/server-fiber"
)

type Option func(*options)

type options struct {
	bindAddress    string
	initializer    srvfiber.Initializer
	timeout        time.Duration
	healthCheckers map[string]Checker
	readyCheckers  map[string]Checker
}

func defaultOpts() options {
	return options{
		bindAddress:    "localhost:8082",
		timeout:        time.Second * 15,
		healthCheckers: make(map[string]Checker),
		readyCheckers:  make(map[string]Checker),
	}
}

func (o *options) GetBindAddress() string {
	return o.bindAddress
}

func WithBindAddress(bindAddress string) Option {
	return func(o *options) {
		o.bindAddress = bindAddress
	}
}

func WithInitializer(initializer srvfiber.Initializer) Option {
	return func(o *options) {
		o.initializer = initializer
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

func WithCheck(name string, checker Checker) Option {
	return func(o *options) {
		o.healthCheckers[name] = checker
	}
}

func WithReadyCheck(name string, checker Checker) Option {
	return func(o *options) {
		o.readyCheckers[name] = checker
	}
}
