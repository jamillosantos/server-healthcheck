//go:generate go run github.com/golang/mock/mockgen -package hcfiber -destination=mocks_mock_test.go . Healthchecker,FiberApp

package hcfiber

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	svchealthcheck "github.com/jamillosantos/services-healthcheck"
)

func TestFiberInitialize(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHC := NewMockHealthchecker(ctrl)

	fiberApp := fiber.New()

	mockHC.EXPECT().Health(gomock.Any()).Return(&svchealthcheck.CheckResponse{
		StatusCode: http.StatusOK,
		Status:     "healthy",
	})

	mockHC.EXPECT().Ready(gomock.Any()).Return(&svchealthcheck.CheckResponse{
		StatusCode: http.StatusCreated,
		Status:     "ready",
	})

	FiberInitialize(mockHC, fiberApp)

	healthCheckResponseWriter, err := fiberApp.Test(httptest.NewRequest("GET", svchealthcheck.HealthPath, nil))
	require.NoError(t, err)
	var healthCheckResponse svchealthcheck.CheckResponse
	err = json.NewDecoder(healthCheckResponseWriter.Body).Decode(&healthCheckResponse)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, healthCheckResponseWriter.StatusCode)
	assert.Equal(t, "healthy", healthCheckResponse.Status)

	readyCheckResponseWriter, err := fiberApp.Test(httptest.NewRequest("GET", svchealthcheck.ReadyPath, nil))
	require.NoError(t, err)
	var readyCheckResponse svchealthcheck.CheckResponse
	err = json.NewDecoder(readyCheckResponseWriter.Body).Decode(&readyCheckResponse)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, readyCheckResponseWriter.StatusCode)
	assert.Equal(t, "ready", readyCheckResponse.Status)
}
