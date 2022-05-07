//go:generate go run github.com/golang/mock/mockgen -package hchttp -destination=mocks_mock_test.go . Healthchecker,ServeMux

package hchttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	svchealthcheck "github.com/jamillosantos/services-healthcheck"
)

func TestHttpInitialize(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockHC := NewMockHealthchecker(ctrl)
	mockSM := NewMockServeMux(ctrl)

	var (
		healthCheck http.Handler
		readyCheck  http.Handler
	)

	mockSM.EXPECT().
		HandleFunc(svchealthcheck.HealthPath, gomock.Any()).
		Do(func(_ string, hc http.HandlerFunc) {
			healthCheck = hc
		})
	mockSM.EXPECT().
		HandleFunc(svchealthcheck.ReadyPath, gomock.Any()).
		Do(func(_ string, hc http.HandlerFunc) {
			readyCheck = hc
		})

	mockHC.EXPECT().Health(ctx).Return(&svchealthcheck.CheckResponse{
		StatusCode: http.StatusOK,
		Status:     "healthy",
	})

	mockHC.EXPECT().Ready(ctx).Return(&svchealthcheck.CheckResponse{
		StatusCode: http.StatusCreated,
		Status:     "ready",
	})

	HttpInitialize(mockHC, mockSM)

	healthCheckResponseWriter := httptest.NewRecorder()
	healthCheck.ServeHTTP(healthCheckResponseWriter, httptest.NewRequest("GET", svchealthcheck.HealthPath, nil))
	var healthCheckResponse svchealthcheck.CheckResponse
	err := json.NewDecoder(healthCheckResponseWriter.Body).Decode(&healthCheckResponse)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, healthCheckResponseWriter.Code)
	assert.Equal(t, "healthy", healthCheckResponse.Status)

	readyCheckResponseWriter := httptest.NewRecorder()
	readyCheck.ServeHTTP(readyCheckResponseWriter, httptest.NewRequest("GET", svchealthcheck.ReadyPath, nil))
	var readyCheckResponse svchealthcheck.CheckResponse
	err = json.NewDecoder(readyCheckResponseWriter.Body).Decode(&readyCheckResponse)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, readyCheckResponseWriter.Code)
	assert.Equal(t, "ready", readyCheckResponse.Status)
}
