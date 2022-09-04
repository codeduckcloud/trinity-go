package container

import (
	"context"
	"fmt"
	"reflect"
	"trinity/core/logx"
)

// DiSelfCheck
// check if the registered instance is invalid
func (s *Container) DiSelfCheck(ctx context.Context, instanceName InstanceName) error {
	var instance interface{}
	switch s.c.InstanceType {
	case MultiInstance:
		pool, ok := s.poolMap[instanceName]
		if !ok {
			return fmt.Errorf("instance `%v` not exist in pool map", instanceName)
		}
		instance = pool.Get()
		defer pool.Put(instance)
	default:
		instance = s.instanceMap[instanceName]
	}

	t := reflect.TypeOf(instance)
	switch t.Kind() {
	case reflect.Ptr:
		instanceVal := reflect.Indirect(reflect.ValueOf(instance))
		for index := 0; index < instanceVal.NumField(); index++ {
			objectName := encodeObjectName(instance, index)
			if _, exist := getTagByName(instance, index, s.c.JsonTagKeyword); !exist {
				logx.FromCtx(ctx).Debugf("%20v: instanceName: %v index: %v objectName: %v, the container tag not exist, skip inject", "di self check", instanceName, index, objectName)
				continue
			}
			resourceName, exist := getStringTagFromContainerByKey(instance, index, s.c.JsonTagKeyword, s.c.ResourceKeyword)
			if !exist {
				return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, the resource tag not exist in container", instanceName, index, objectName)
			}
			val := instanceVal.Field(index)
			autoWire := s.getAutoWireTag(instance, index)
			if !autoWire {
				if val.CanSet() {
					logx.FromCtx(ctx).Warnf("self check warning: instanceName: %v index: %v objectName: %v, auto wired is false but the param can be injected ", instanceName, index, objectName)
				}
				continue
			}
			if !val.CanSet() {
				return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, private param", instanceName, index, objectName)
			}
			if !val.IsZero() {
				return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, the param to be injected is not null", instanceName, index, objectName)
			}
			switch s.c.InstanceType {
			case MultiInstance:
				instancePool, exist := s.poolMap[InstanceName(resourceName)]
				if !exist {
					return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, resource name: %v not register in container ", instanceName, index, objectName, resourceName)
				}
				switch val.Kind() {
				case reflect.Interface:
					instance := instancePool.Get()
					defer instancePool.Put(instance)
					instanceType := reflect.TypeOf(instance)
					if !instanceType.Implements(val.Type()) {
						return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, resource name: %v type: %v not implement the interface %v", instanceName, index, objectName, resourceName, instanceType.Name(), val.Type().Name())
					}
				default:
					instance := instancePool.Get()
					defer instancePool.Put(instance)
					instanceType := reflect.TypeOf(instance)
					if val.Type() != instanceType {
						return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, resource name: %v type not same, expected: %v actual: %v", instanceName, index, objectName, resourceName, val.Type(), instanceType)
					}
				}
			default:
				instance, exist := s.instanceMap[InstanceName(resourceName)]
				if !exist {
					return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, resource name: %v not register in container ", instanceName, index, objectName, resourceName)
				}
				instanceType := reflect.TypeOf(instance)
				switch val.Kind() {
				case reflect.Interface:
					if !instanceType.Implements(val.Type()) {
						return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, resource name: %v type: %v not implement the interface %v", instanceName, index, objectName, resourceName, instanceType.Name(), val.Type().Name())
					}
				default:
					if val.Type() != instanceType {
						return fmt.Errorf("self check error: instanceName: %v index: %v objectName: %v, resource name: %v type not same, expected: %v actual: %v", instanceName, index, objectName, resourceName, val.Type(), instanceType)
					}
				}

			}

		}
		return nil
	default:
		return fmt.Errorf("the object to be injected %v should be addressable", t)
	}
}

func (s *Container) DiAllFields(ctx context.Context, dest interface{}, injectingMap map[InstanceName]interface{}) {
	destVal := reflect.Indirect(reflect.ValueOf(dest))
	for index := 0; index < destVal.NumField(); index++ {
		if _, exist := getTagByName(dest, index, s.c.JsonTagKeyword); !exist {
			continue
		}
		resourceName, _ := getStringTagFromContainerByKey(dest, index, s.c.JsonTagKeyword, s.c.ResourceKeyword)
		val := destVal.Field(index)
		autoWire := s.getAutoWireTag(dest, index)
		if !autoWire {
			continue
		}
		if instance, exist := injectingMap[InstanceName(resourceName)]; exist {
			val.Set(reflect.ValueOf(instance))
			continue
		}
		instance := s.GetInstance(ctx, InstanceName(resourceName), injectingMap)
		val.Set(reflect.ValueOf(instance))
	}
}

func (s *Container) DiFree(ctx context.Context, dest interface{}) {
	t := reflect.TypeOf(dest)
	switch t.Kind() {
	case reflect.Ptr:
		destVal := reflect.Indirect(reflect.ValueOf(dest))
		for index := 0; index < destVal.NumField(); index++ {
			objectName := encodeObjectName(dest, index)
			if _, exist := getTagByName(dest, index, s.c.JsonTagKeyword); !exist {
				logx.FromCtx(ctx).Debugf("objectName di free skipped => %v, container not exist", objectName)
				continue
			}
			val := destVal.Field(index)
			autoWire := s.getAutoWireTag(dest, index)
			if !autoWire {
				continue
			}
			if val.CanSet() {
				val.Set(reflect.Zero(val.Type()))
			}
		}
	}
}
