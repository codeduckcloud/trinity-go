package container

import (
	"github.com/codeduckcloud/trinity-go/core/logx"
)

var (
	logger     = logx.NewLogrusLogger()
	logWithCtx = logx.NewCtx(logger)
)
