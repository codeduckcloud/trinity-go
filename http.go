package trinity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"sync"
	"syscall"

	"github.com/codeduckcloud/trinity-go/core/container"
	"github.com/codeduckcloud/trinity-go/core/httpx"
	"github.com/codeduckcloud/trinity-go/core/logx"
	"github.com/codeduckcloud/trinity-go/middleware"
)

var (
	// booting instance
	_bootingInstances      []bootingInstance
	_bootingMultiInstances []bootingMultiInstance
	_bootingControllers    []bootingController
	// booting cache
	injectMapPool = &sync.Pool{
		New: func() interface{} {
			return make(map[container.InstanceName]interface{})
		},
	}
)

type bootingController struct {
	rootPath     string
	instanceName container.InstanceName
	requestMaps  []RequestMap
}

type RequestMap struct {
	method   string
	subPath  string
	funcName string
	handlers []func(http.Handler) http.Handler
	isRaw    bool
}

type bootingInstance struct {
	instanceName container.InstanceName
	instance     interface{}
}

type bootingMultiInstance struct {
	instanceName container.InstanceName
	instancePool *sync.Pool
}

func RegisterInstance(instanceName container.InstanceName, instance interface{}) {
	newInstance := bootingInstance{
		instanceName: instanceName,
		instance:     instance,
	}
	_bootingInstances = append(_bootingInstances, newInstance)
}

func RegisterMultiInstance(instanceName container.InstanceName, instancePool *sync.Pool) {
	newInstance := bootingMultiInstance{
		instanceName: instanceName,
		instancePool: instancePool,
	}
	_bootingMultiInstances = append(_bootingMultiInstances, newInstance)
}

func RegisterController(rootPath string, instanceName container.InstanceName, requestMaps ...RequestMap) {
	newController := bootingController{
		rootPath:     rootPath,
		instanceName: instanceName,
		requestMaps:  requestMaps,
	}
	_bootingControllers = append(_bootingControllers, newController)
}

func NewRequestMapping(method string, path string, funcName string, handlers ...func(http.Handler) http.Handler) RequestMap {
	return RequestMap{
		method:   method,
		subPath:  path,
		funcName: funcName,
		handlers: handlers,
		isRaw:    false,
	}
}

func NewRawRequestMapping(method string, path string, funcName string, handlers ...func(http.Handler) http.Handler) RequestMap {
	return RequestMap{
		method:   method,
		subPath:  path,
		funcName: funcName,
		handlers: handlers,
		isRaw:    true,
	}
}

func (t *trinity) initInstance(ctx context.Context) {
	switch t.container.GetInstanceType() {
	case container.MultiInstance:
		for _, instance := range _bootingMultiInstances {
			t.container.RegisterMultiInstance(ctx, instance.instanceName, instance.instancePool)
			logx.FromCtx(ctx).Infof("%-8v %-10v %-7v => %v ", "instance", "register", "success", instance.instanceName)
		}
		if err := t.container.InstanceDISelfCheck(ctx); err != nil {
			logx.FromCtx(ctx).Fatalf("%-10v %-10v %-7v, err: %v", "instance", "self-check", "failed", err)
		}
	default:
		for _, instance := range _bootingInstances {
			t.container.RegisterInstance(ctx, instance.instanceName, instance.instance)
			logx.FromCtx(ctx).Infof("%-8v %-10v %-7v => %v ", "instance", "register", "success", instance.instanceName)
		}
		if err := t.container.InstanceDISelfCheck(ctx); err != nil {
			logx.FromCtx(ctx).Fatalf("%-10v %-10v %-7v, err: %v", "instance", "self-check", "failed", err)
		}
	}

}

func (t *trinity) diRouter(ctx context.Context) {
	t.mux.Use(logx.SessionLogger(ctx))
	t.mux.Use(middleware.Recovery())
	t.routerSelfCheck(ctx)
	// register router
	for _, controller := range _bootingControllers {
		for _, requestMapping := range controller.requestMaps {
			urlPath := filepath.Join(controller.rootPath, requestMapping.subPath)
			h := http.HandlerFunc(DIHandler(t.container, controller.instanceName, requestMapping.funcName, requestMapping.isRaw))
			for i := len(requestMapping.handlers) - 1; i >= 0; i-- {
				h = requestMapping.handlers[i](h).ServeHTTP
			}
			t.mux.MethodFunc(requestMapping.method, urlPath, h)
			logx.FromCtx(ctx).Infof("router   register handler: %-6s %-30s => %v.%v ", requestMapping.method, urlPath, controller.instanceName, requestMapping.funcName)
		}
	}
}

func (t *trinity) routerSelfCheck(ctx context.Context) {
	for _, controller := range _bootingControllers {
		for _, requestMap := range controller.requestMaps {
			injectMap := injectMapPool.Get().(map[container.InstanceName]interface{})
			instance := t.container.GetInstance(ctx, controller.instanceName, injectMap)
			defer func() {
				for k, v := range injectMap {
					t.container.Release(ctx, k, v)
					delete(injectMap, k)
				}
				injectMapPool.Put(injectMap)
			}()
			_, ok := reflect.TypeOf(instance).MethodByName(requestMap.funcName)
			if !ok {
				logx.FromCtx(ctx).Fatalf("%-8v %-10v %-7v => %v.%v , func %v not exist ", "router", "self-check", "failed", controller.instanceName, requestMap.funcName, requestMap.funcName)
				continue
			}
			logx.FromCtx(ctx).Infof("%-8v %-10v %-7v => %v.%v ", "router", "self-check", "success", controller.instanceName, requestMap.funcName)
		}
	}

}

// multi instance di handler
func DIHandler(c *container.Container, instanceName container.InstanceName, funcName string, isRaw bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), httpx.HttpxContext, httpx.NewContext(r, 0)))
		injectMap := injectMapPool.Get().(map[container.InstanceName]interface{})
		instance := c.GetInstance(r.Context(), instanceName, injectMap)
		defer func() {
			for k, v := range injectMap {
				c.Release(r.Context(), k, v)
				delete(injectMap, k)
			}
			injectMapPool.Put(injectMap)
		}()
		currentMethod, ok := reflect.TypeOf(instance).MethodByName(funcName)
		if !ok {
			panic("method not registered, please ensure your run thee RouterSelfCheck before start your service")
		}
		inParams, err := httpx.InvokeMethod(currentMethod.Type, r, instance, w)
		if err != nil {
			httpx.HttpResponseErr(r.Context(), w, err)
			return
		}
		responseValue := currentMethod.Func.Call(inParams)
		if isRaw {
			return
		}
		switch len(responseValue) {
		case 0:
			httpx.HttpResponse(r.Context(), w, httpx.GetHTTPStatusCode(r.Context(), httpx.DefaultHttpSuccessCode), nil)
			return
		case 1:
			if err, ok := responseValue[0].Interface().(error); ok {
				if err != nil {
					httpx.HttpResponseErr(r.Context(), w, err)
					return
				}
			}
			httpx.HttpResponse(r.Context(), w, httpx.GetHTTPStatusCode(r.Context(), httpx.DefaultHttpSuccessCode), responseValue[0].Interface())
			return
		case 2:
			if err, ok := responseValue[1].Interface().(error); ok {
				if err != nil {
					httpx.HttpResponseErr(r.Context(), w, err)
					return
				}
			}
			httpx.HttpResponse(r.Context(), w, httpx.GetHTTPStatusCode(r.Context(), httpx.DefaultHttpSuccessCode), responseValue[0].Interface())
			return
		default:
			httpx.HttpResponseErr(r.Context(), w, errors.New("wrong res type , first out should be response value , second out should be error "))
			return
		}
	}
}

func (t *trinity) ServeHTTP(ctx context.Context, addr ...string) error {
	address := ":http"
	if len(addr) > 0 {
		address = addr[0]
	}
	logx.FromCtx(ctx).Infof("http service started at %v", address)
	gErr := make(chan error)
	go func() {
		gErr <- http.ListenAndServe(address, t.mux)
	}()
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		gErr <- fmt.Errorf("receive %s", <-sigChan)
	}()

	return <-gErr
}
