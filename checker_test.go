package svchealthcheck

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckerFunc_Check(t *testing.T) {
	wantCtx := context.Background()
	wantErr := errors.New("test error")
	called := false
	checker := CheckerFunc(func(ctx context.Context) error {
		assert.Equal(t, wantCtx, ctx)
		called = true
		return wantErr
	})
	gotErr := checker.Check(wantCtx)
	require.True(t, called, "the Check method was never called")
	assert.Equal(t, wantErr, gotErr)

}
