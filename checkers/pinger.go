package checkers

import (
	"context"

	srvhealthcheck "github.com/jamillosantos/services-healthcheck"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func PingerChecker(pinger Pinger) srvhealthcheck.Checker {
	return srvhealthcheck.CheckerFunc(func(ctx context.Context) error {
		return pinger.Ping(ctx)
	})
}
