//go:generate go run github.com/golang/mock/mockgen -package svchealthcheck -destination=locker_mock_test.go sync Locker
//go:generate go run github.com/golang/mock/mockgen -package svchealthcheck -destination=checker_mock_test.go . Checker

package svchealthcheck

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_panicErrorMessage(t *testing.T) {
	t.Run("should return a string given a non-error", func(t *testing.T) {
		got := panicError("something wrong happened")
		assert.Equal(t, "checker panicked: something wrong happened", got.Error())
	})

	t.Run("should return a string given an error", func(t *testing.T) {
		got := panicError(errors.New("something wrong happened"))
		assert.Equal(t, "checker panicked: something wrong happened", got.Error())
	})
}

func Test_errorMessage(t *testing.T) {
	t.Run("should return an empty string if the given error is nil", func(t *testing.T) {
		got := errorMessage(nil)
		assert.Empty(t, got)
	})

	t.Run("should return a string given an error", func(t *testing.T) {
		wantMsg := "something wrong happened"
		got := errorMessage(errors.New(wantMsg))
		assert.Equal(t, wantMsg, got)
	})
}

func Test_handlerRecover(t *testing.T) {
	t.Run("should do nothing if no panic happened", func(t *testing.T) {
		handlerRecover(nil, nil)
	})

	t.Run("should set an error if a panic happened", func(t *testing.T) {
		err := errors.New("something wrong happened")

		errch := make(chan error, 1)
		handlerRecover(err, errch)

		select {
		case err = <-errch:
			assert.Contains(t, err.Error(), ErrCheckerPanic.Error())
		default:
			require.Fail(t, "should have received an error")
			return
		}
	})
}

func TestHealthcheck_generateResponse(t *testing.T) {
	wantDuration1 := time.Millisecond * 500
	wantDuration2 := time.Millisecond * 300
	wantDuration3 := time.Millisecond * 750

	t.Run("should return ok when all checks pass", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockChecker1 := NewMockChecker(ctrl)
		mockChecker2 := NewMockChecker(ctrl)
		mockChecker3 := NewMockChecker(ctrl)

		mockChecker1.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration1)
			}).
			Return(nil)

		mockChecker2.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration2)
			}).Return(nil)

		mockChecker3.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration3)
			}).
			Return(nil)

		hc := NewHealthcheck()

		response := hc.generateResponse(context.Background(), map[string]Checker{
			"check1": mockChecker1,
			"check2": mockChecker2,
			"check3": mockChecker3,
		})

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "OK", response.Status)
		assert.Len(t, response.Checks, 3)
		assert.Empty(t, response.Checks["check1"].Error)
		assert.InDelta(t, wantDuration1, mustParseDuration(t, response.Checks["check1"].Duration), float64(time.Millisecond*25))
		assert.Empty(t, response.Checks["check2"].Error)
		assert.InDelta(t, wantDuration2, mustParseDuration(t, response.Checks["check2"].Duration), float64(time.Millisecond*25))
		assert.Empty(t, response.Checks["check3"].Error)
		assert.InDelta(t, wantDuration3, mustParseDuration(t, response.Checks["check3"].Duration), float64(time.Millisecond*25))
	})

	t.Run("should return internal server error when a checker panics", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockChecker1 := NewMockChecker(ctrl)
		mockChecker2 := NewMockChecker(ctrl)
		mockChecker3 := NewMockChecker(ctrl)

		mockChecker1.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration1)
			}).
			Return(nil)

		mockChecker2.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration2)

				panic(errors.New("panicked"))
			}).Return(nil)

		mockChecker3.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration3)
			}).
			Return(errors.New("some error"))

		hc := NewHealthcheck()

		response := hc.generateResponse(context.Background(), map[string]Checker{
			"check1": mockChecker1,
			"check2": mockChecker2,
			"check3": mockChecker3,
		})

		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "Internal Server Error", response.Status)
		assert.Len(t, response.Checks, 3)
		assert.Empty(t, response.Checks["check1"].Error)
		assert.InDelta(t, wantDuration1, mustParseDuration(t, response.Checks["check1"].Duration), float64(time.Millisecond*25))
		assert.Equal(t, response.Checks["check2"].Error, "checker panicked: panicked")
		assert.InDelta(t, wantDuration2, mustParseDuration(t, response.Checks["check2"].Duration), float64(time.Millisecond*25))
		assert.Equal(t, response.Checks["check3"].Error, "some error")
		assert.InDelta(t, wantDuration3, mustParseDuration(t, response.Checks["check3"].Duration), float64(time.Millisecond*25))
	})

	t.Run("should return service unavailable when one check fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockChecker1 := NewMockChecker(ctrl)
		mockChecker2 := NewMockChecker(ctrl)
		mockChecker3 := NewMockChecker(ctrl)

		mockChecker1.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration1)
			}).
			Return(nil)

		mockChecker2.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration2)
			}).Return(nil)

		mockChecker3.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration3)
			}).
			Return(errors.New("some error"))

		hc := NewHealthcheck()

		response := hc.generateResponse(context.Background(), map[string]Checker{
			"check1": mockChecker1,
			"check2": mockChecker2,
			"check3": mockChecker3,
		})

		assert.Equal(t, http.StatusServiceUnavailable, response.StatusCode)
		assert.Equal(t, "Service Unavailable", response.Status)
		assert.Len(t, response.Checks, 3)
		assert.Empty(t, response.Checks["check1"].Error)
		assert.InDelta(t, wantDuration1, mustParseDuration(t, response.Checks["check1"].Duration), float64(time.Millisecond*25))
		assert.Empty(t, response.Checks["check2"].Error)
		assert.InDelta(t, wantDuration2, mustParseDuration(t, response.Checks["check2"].Duration), float64(time.Millisecond*25))
		assert.Equal(t, response.Checks["check3"].Error, "some error")
		assert.InDelta(t, wantDuration3, mustParseDuration(t, response.Checks["check3"].Duration), float64(time.Millisecond*25))
	})

	t.Run("should return service unavailable when one Checker times out", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockChecker1 := NewMockChecker(ctrl)
		mockChecker2 := NewMockChecker(ctrl)
		mockChecker3 := NewMockChecker(ctrl)

		mockChecker1.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration1)
			}).
			Return(nil)

		mockChecker2.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration2)
			}).Return(nil)

		mockChecker3.EXPECT().
			Check(gomock.Any()).
			Do(func(_ context.Context) {
				time.Sleep(wantDuration3)
			}).
			Return(errors.New("some error"))

		wantTimeout := wantDuration1 + time.Millisecond*50
		hc := NewHealthcheck(WithTimeout(wantTimeout))

		response := hc.generateResponse(context.Background(), map[string]Checker{
			"check1": mockChecker1,
			"check2": mockChecker2,
			"check3": mockChecker3,
		})

		assert.Equal(t, http.StatusServiceUnavailable, response.StatusCode)
		assert.Equal(t, "Service Unavailable", response.Status)
		assert.Len(t, response.Checks, 3)
		assert.Empty(t, response.Checks["check1"].Error)
		assert.InDelta(t, wantDuration1, mustParseDuration(t, response.Checks["check1"].Duration), float64(time.Millisecond*25))
		assert.Empty(t, response.Checks["check2"].Error)
		assert.InDelta(t, wantDuration2, mustParseDuration(t, response.Checks["check2"].Duration), float64(time.Millisecond*25))
		assert.Equal(t, response.Checks["check3"].Error, context.DeadlineExceeded.Error())
		assert.InDelta(t, wantTimeout, mustParseDuration(t, response.Checks["check3"].Duration), float64(time.Millisecond*25))
	})
}

func mustParseDuration(t *testing.T, s string) time.Duration {
	t.Helper()
	d, err := time.ParseDuration(s)
	require.NoError(t, err)
	return d
}

func TestHealthcheck_Health(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChecker1 := NewMockChecker(ctrl)
	mockChecker2 := NewMockChecker(ctrl)
	mockChecker3 := NewMockChecker(ctrl)

	mockChecker1.EXPECT().
		Check(gomock.Any()).
		Return(nil)

	mockChecker2.EXPECT().
		Check(gomock.Any()).
		Return(nil)

	mockChecker3.EXPECT().
		Check(gomock.Any()).
		Return(nil)

	hc := NewHealthcheck(
		WithCheck("check1", mockChecker1),
		WithCheck("check2", mockChecker2),
		WithCheck("check3", mockChecker3),
	)

	response := hc.Health(context.Background())

	assert.Len(t, response.Checks, 3)
	assert.Empty(t, response.Checks["check1"].Error)
	assert.Empty(t, response.Checks["check2"].Error)
	assert.Empty(t, response.Checks["check3"].Error)
}

func TestHealthcheck_Ready(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChecker1 := NewMockChecker(ctrl)
	mockChecker2 := NewMockChecker(ctrl)
	mockChecker3 := NewMockChecker(ctrl)

	mockChecker1.EXPECT().
		Check(gomock.Any()).
		Return(nil)

	mockChecker2.EXPECT().
		Check(gomock.Any()).
		Return(nil)

	mockChecker3.EXPECT().
		Check(gomock.Any()).
		Return(nil)

	hc := NewHealthcheck(
		WithReadyCheck("check1", mockChecker1),
		WithReadyCheck("check2", mockChecker2),
		WithReadyCheck("check3", mockChecker3),
	)

	response := hc.Ready(context.Background())

	assert.Len(t, response.Checks, 3)
	assert.Empty(t, response.Checks["check1"].Error)
	assert.Empty(t, response.Checks["check2"].Error)
	assert.Empty(t, response.Checks["check3"].Error)
}
