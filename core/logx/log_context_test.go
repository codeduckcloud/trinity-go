package logx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogContext(t *testing.T) {
	logger := NewLogrusLogger()
	ctx := NewCtx(logger)
	got := FromCtx(ctx)
	assert.Equal(t, logger, got)

	ctx2 := WithCtx(context.Background(), logger)
	assert.Equal(t, logger, FromCtx(ctx2))
}

func TestFromCtx_Panic(t *testing.T) {
	assert.Panics(t, func() {
		FromCtx(context.Background())
	})
}
