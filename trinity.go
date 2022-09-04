package trinity

import (
	"context"

	"github.com/codeduckcloud/trinity-go/core/container"

	"github.com/go-chi/chi/v5"
)

var (
	_defaultRouter = chi.NewRouter()
)

type Config struct {
	Mux          mux
	InstanceType container.InstanceType
}

type trinity struct {
	mux
	container *container.Container
}

func New(ctx context.Context, c ...Config) *trinity {
	if len(c) > 0 {
		if c[0].Mux == nil {
			c[0].Mux = _defaultRouter
		}
		if c[0].InstanceType == "" {
			c[0].InstanceType = container.Singleton
		}
	} else {
		c = append(c, Config{
			Mux:          _defaultRouter,
			InstanceType: container.Singleton,
		})
	}
	ins := &trinity{
		mux:       c[0].Mux,
		container: container.NewContainer(),
	}
	ins.initInstance(ctx)
	ins.diRouter(ctx)
	return ins
}
