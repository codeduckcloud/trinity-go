package controller

import (
	"github.com/codeduckcloud/trinity-go"
)

func init() {
	trinity.RegisterInstance("BenchmarkController", &benchmarkControllerImpl{})
	trinity.RegisterController("/benchmark", "BenchmarkController",
		trinity.NewRequestMapping("GET", "/simple", "Simple"),
		trinity.NewRequestMapping("GET", "/simple/{id}", "PathParam"),
	)
}

type benchmarkControllerImpl struct {
}

func (c *benchmarkControllerImpl) Simple() string {
	return "ok"
}

func (c *benchmarkControllerImpl) PathParam(Args struct {
	ID int `path_param:"id"`
}) int {
	return Args.ID
}
