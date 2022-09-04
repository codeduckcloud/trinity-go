package container

import (
	"context"
	"reflect"
	"sync"

	"github.com/codeduckcloud/trinity-go/core/logx"
)

type Config struct {
	// the AutoWire flag here decide all the param inside the container should be injected
	// if multi instance and auto wire is true, the value will be set to zero after request finished
	// default value: true
	AutoWire bool
	// the json tag name will read from struct, support to customize
	// default value: container
	JsonTagKeyword Keyword
	// the value tag name will read the auto wire value, support to customize
	// default value: autowire
	AutoWireKeyword Keyword
	// the resource tag name will read the resource name, support to customize
	// default value: resource
	ResourceKeyword Keyword
	InstanceType    InstanceType
}

var (
	// the default config for container
	_DefaultConfig = &Config{
		AutoWire:        true,
		JsonTagKeyword:  _CONTAINER,
		AutoWireKeyword: _AUTO_WIRE,
		ResourceKeyword: _RESOURCE,
		InstanceType:    Singleton,
	}
)

type Container struct {
	c *Config
	// multi instance
	// pool map
	poolMap map[InstanceName]*sync.Pool
	// poolTypeMap caching the type info
	poolTypeMap map[InstanceName]reflect.Type

	// singleton
	// instance pool map
	instanceMap            map[InstanceName]interface{}
	instanceMapInitialized map[InstanceName]interface{}
}

// NewContainer get the new container instance
// if not passing the config , will init with the default config
func NewContainer(c ...Config) *Container {
	newContainer := new(Container)
	newContainer.poolMap = make(map[InstanceName]*sync.Pool)
	newContainer.poolTypeMap = make(map[InstanceName]reflect.Type)
	newContainer.instanceMap = make(map[InstanceName]interface{})
	newContainer.instanceMapInitialized = make(map[InstanceName]interface{})
	if len(c) > 0 {
		if c[0].JsonTagKeyword == "" {
			c[0].JsonTagKeyword = _CONTAINER
		}
		if c[0].AutoWireKeyword == "" {
			c[0].AutoWireKeyword = _AUTO_WIRE
		}
		if c[0].ResourceKeyword == "" {
			c[0].ResourceKeyword = _RESOURCE
		}
		if c[0].InstanceType == "" {
			c[0].InstanceType = Singleton
		}
		newContainer.c = &c[0]
	} else {
		newContainer.c = _DefaultConfig
	}
	return newContainer
}

// RegisterInstance
// register singleton instance
// golang using ctx to pass the session related data, so
// we are using singleton instance by default
func (s *Container) RegisterInstance(ctx context.Context, instanceName InstanceName, instance interface{}) {
	if s.c.InstanceType != Singleton {
		logx.FromCtx(ctx).Fatal("cannot register multi instance in multi instance mode")
	}
	if err := instanceName.Validate(ctx); err != nil {
		logx.FromCtx(ctx).Fatal(err)
	}
	if instance == nil {
		logx.FromCtx(ctx).Fatal("instance pool cannot be empty")
	}
	if _, exist := s.instanceMap[instanceName]; exist {
		logx.FromCtx(ctx).Fatalf("instance name %v already existed, cannot register instance with the same name ", instanceName)
	}
	s.instanceMap[instanceName] = instance
}

// RegisterMultiInstance register new multi instance
// the multi instance will create the new struct when new request coming
// if instanceName is empty will fatal
// if instancePool is invalid , will fatal
func (s *Container) RegisterMultiInstance(ctx context.Context, instanceName InstanceName, instancePool *sync.Pool) {
	if s.c.InstanceType != MultiInstance {
		logx.FromCtx(ctx).Fatal("cannot register multi instance in singleton instance mode")
	}
	if err := instanceName.Validate(ctx); err != nil {
		logx.FromCtx(ctx).Fatal(err)
	}
	if instancePool == nil {
		logx.FromCtx(ctx).Fatal("instance pool cannot be empty")
	}
	if _, ok := s.poolMap[instanceName]; ok {
		logx.FromCtx(ctx).Fatalf("instance name %v already existed, cannot register instance with the same name ", instanceName)
	}
	ins := instancePool.Get()
	defer instancePool.Put(ins)
	t := reflect.TypeOf(ins)
	s.poolMap[instanceName] = instancePool
	s.poolTypeMap[instanceName] = t
}

// CheckInstanceNameIfExist
// check instance name if exist
// if exist , return true
// if not exist , return false
func (s *Container) CheckInstanceNameIfExist(instanceName InstanceName) bool {
	_, ok := s.poolMap[instanceName]
	return ok
}

// InstanceDISelfCheck
// self check all the instance registered exist or not
func (s *Container) InstanceDISelfCheck(ctx context.Context) error {
	switch s.c.InstanceType {
	case MultiInstance:
		for k := range s.poolMap {
			if err := s.DiSelfCheck(ctx, k); err != nil {
				logx.FromCtx(ctx).Errorf("%-8v %-10v %-7v => %v, error: %v", "instance", "self-check", "failed", k, err)
				return err
			}
			logx.FromCtx(ctx).Infof("%-8v %-10v %-7v => %v", "instance", "self-check", "success", k)
		}
	default:
		for k := range s.instanceMap {
			if err := s.DiSelfCheck(ctx, k); err != nil {
				logx.FromCtx(ctx).Errorf("%-8v %-10v %-7v => %v, error: %v", "instance", "self-check", "failed", k, err)
				return err
			}
			service := s.GetInstance(ctx, k, s.instanceMapInitialized)
			s.instanceMapInitialized[k] = service
		}
	}

	return nil
}

// InstanceDISelfCheck
// get instance by instance name
// injectingMap , the dependency instance, will inject the instance in injectingMap as priority
func (s *Container) GetInstance(ctx context.Context, instanceName InstanceName, injectingMap map[InstanceName]interface{}) interface{} {
	if v, ok := injectingMap[instanceName]; ok {
		return v
	}
	switch s.c.InstanceType {
	case MultiInstance:
		pool, ok := s.poolMap[instanceName]
		if !ok {
			logx.FromCtx(ctx).Panicf("instance not exist in container => %v", instanceName)
		}
		service := pool.Get()
		injectingMap[instanceName] = service
		s.DiAllFields(ctx, service, injectingMap)
		return service
	default:
		service, ok := s.instanceMapInitialized[instanceName]
		if ok {
			return service
		}
		s.DiAllFields(ctx, s.instanceMap[instanceName], injectingMap)
		s.instanceMapInitialized[instanceName] = s.instanceMap[instanceName]
		return s.instanceMapInitialized[instanceName]
	}
}

// Release
// release the instance to instance pool
func (s *Container) Release(ctx context.Context, instanceName InstanceName, instance interface{}) {
	if s.c.InstanceType == Singleton {
		return
	}
	instancePool, ok := s.poolMap[instanceName]
	if !ok {
		logx.FromCtx(ctx).Errorf("instance release failed => %v, not exist in container", instanceName)
		return
	}
	if reflect.TypeOf(instance) != s.poolTypeMap[instanceName] {
		logx.FromCtx(ctx).Errorf("released wrong types instance to instance pool")
		return
	}
	s.DiFree(ctx, instance)
	instancePool.Put(instance)
}

func (s *Container) getAutoWireTag(obj interface{}, index int) bool {
	v, exist := getBoolTagFromContainer(obj, index, s.c.JsonTagKeyword, s.c.AutoWireKeyword)
	if exist {
		return v
	}
	return s.c.AutoWire
}

func (s *Container) GetInstanceType() InstanceType {
	return s.c.InstanceType
}
