package container

import (
	"trinity/core/logx"
)

var (
	logger     = logx.NewLogrusLogger()
	logWithCtx = logx.NewCtx(logger)
)
