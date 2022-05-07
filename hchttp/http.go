package hchttp

import (
	"context"
	"encoding/json"
	"net/http"

	svchealthcheck "github.com/jamillosantos/services-healthcheck"
)

// Healthchecker abstracts the implementation of the svchealthcheck.Healthcheck.
type Healthchecker interface {
	Health(ctx context.Context) *svchealthcheck.CheckResponse
	Ready(ctx context.Context) *svchealthcheck.CheckResponse
}

// ServeMux abstracts the implementation of the http.ServeMux.
type ServeMux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// HttpInitialize that will set up the endpoints on a given http.ServeMux.
func HttpInitialize(healthcheck Healthchecker, mux ServeMux) {
	mux.HandleFunc(svchealthcheck.HealthPath, httpEndpoint(healthcheck.Health))
	mux.HandleFunc(svchealthcheck.ReadyPath, httpEndpoint(healthcheck.Ready))
}

func httpEndpoint(getResponse func(ctx context.Context) *svchealthcheck.CheckResponse) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		r := getResponse(request.Context())
		writer.WriteHeader(r.StatusCode)
		_ = json.NewEncoder(writer).Encode(r)
	}
}
