package container

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceName_Validate(t *testing.T) {
	assert.NoError(t, InstanceName("ok").Validate(logWithCtx))
	assert.Error(t, InstanceName("").Validate(logWithCtx))
}

func TestContainer_GetInstanceType(t *testing.T) {
	c := NewContainer()
	assert.Equal(t, Singleton, c.GetInstanceType())
	c2 := NewContainer(Config{InstanceType: MultiInstance})
	assert.Equal(t, MultiInstance, c2.GetInstanceType())
}

func TestNewContainer_WithCustomConfig(t *testing.T) {
	// trigger defaults branches by providing an all-empty Config
	c := NewContainer(Config{})
	assert.Equal(t, _CONTAINER, c.c.JsonTagKeyword)
	assert.Equal(t, _AUTO_WIRE, c.c.AutoWireKeyword)
	assert.Equal(t, _RESOURCE, c.c.ResourceKeyword)
	assert.Equal(t, Singleton, c.c.InstanceType)
}

func TestContainer_RegisterInstance_Fatal_WrongMode(t *testing.T) {
	// with noop ExitFunc, Fatal does not exit and code continues.
	c := NewContainer(Config{InstanceType: MultiInstance})
	// RegisterInstance called on multi-instance container logs fatal and continues.
	assert.NotPanics(t, func() {
		c.RegisterInstance(logWithCtx, "x", &userRep{})
	})
}

func TestContainer_RegisterInstance_Fatal_EmptyName(t *testing.T) {
	c := NewContainer()
	assert.NotPanics(t, func() {
		c.RegisterInstance(logWithCtx, "", &userRep{})
	})
}

func TestContainer_RegisterInstance_Fatal_NilInstance(t *testing.T) {
	c := NewContainer()
	assert.NotPanics(t, func() {
		c.RegisterInstance(logWithCtx, "x", nil)
	})
}

func TestContainer_RegisterInstance_Fatal_AlreadyRegistered(t *testing.T) {
	c := NewContainer()
	c.RegisterInstance(logWithCtx, "x", &userRep{})
	assert.NotPanics(t, func() {
		c.RegisterInstance(logWithCtx, "x", &userRep{})
	})
}

func TestContainer_RegisterMultiInstance_Fatal_WrongMode(t *testing.T) {
	c := NewContainer()
	assert.NotPanics(t, func() {
		c.RegisterMultiInstance(logWithCtx, "x", &sync.Pool{New: func() interface{} { return &userRep{} }})
	})
}

func TestContainer_RegisterMultiInstance_Fatal_EmptyName(t *testing.T) {
	c := NewContainer(Config{InstanceType: MultiInstance})
	assert.NotPanics(t, func() {
		c.RegisterMultiInstance(logWithCtx, "", &sync.Pool{New: func() interface{} { return &userRep{} }})
	})
}

func TestContainer_RegisterMultiInstance_Fatal_NilPool(t *testing.T) {
	c := NewContainer(Config{InstanceType: MultiInstance})
	// without a real os.Exit, the noop Fatal proceeds and the nil pool
	// dereference causes a panic; both outcomes exercise the logged branch.
	assert.Panics(t, func() {
		c.RegisterMultiInstance(logWithCtx, "x", nil)
	})
}

func TestContainer_RegisterMultiInstance_Fatal_AlreadyRegistered(t *testing.T) {
	c := NewContainer(Config{InstanceType: MultiInstance})
	pool := &sync.Pool{New: func() interface{} { return &userRep{} }}
	c.RegisterMultiInstance(logWithCtx, "x", pool)
	assert.NotPanics(t, func() {
		c.RegisterMultiInstance(logWithCtx, "x", pool)
	})
}

func TestContainer_InstanceDISelfCheck_MultiError(t *testing.T) {
	c := NewContainer(Config{InstanceType: MultiInstance})
	c.RegisterMultiInstance(logWithCtx, "bad", &sync.Pool{New: func() interface{} { return testInjectErr1{} }})
	err := c.InstanceDISelfCheck(logWithCtx)
	assert.Error(t, err)
}

func TestContainer_InstanceDISelfCheck_SingletonSuccess(t *testing.T) {
	c := NewContainer(Config{InstanceType: Singleton})
	c.RegisterInstance(logWithCtx, "SingletonSelfCheck1", &SingletonSelfCheck1{})
	c.RegisterInstance(logWithCtx, "SingletonSelfCheck2", &SingletonSelfCheck2{})
	c.RegisterInstance(logWithCtx, "SingletonSelfCheck3", &SingletonSelfCheck3{})
	c.RegisterInstance(logWithCtx, "SingletonSelfCheck4", &singletonSelfCheck4{})
	assert.NoError(t, c.InstanceDISelfCheck(logWithCtx))
}

func TestContainer_InstanceDISelfCheck_SingletonError(t *testing.T) {
	c := NewContainer(Config{InstanceType: Singleton})
	c.RegisterInstance(logWithCtx, "bad", &testInjectErr3{})
	assert.Error(t, c.InstanceDISelfCheck(logWithCtx))
}

func TestContainer_GetInstance_MultiPanicWhenNotExist(t *testing.T) {
	c := NewContainer(Config{InstanceType: MultiInstance})
	assert.Panics(t, func() {
		c.GetInstance(logWithCtx, "missing", make(map[InstanceName]interface{}))
	})
}

func TestContainer_GetInstance_SingletonInitializedCache(t *testing.T) {
	c := NewContainer(Config{InstanceType: Singleton})
	c.RegisterInstance(logWithCtx, "rep", &userRep{})
	// first call initializes and caches
	first := c.GetInstance(logWithCtx, "rep", map[InstanceName]interface{}{})
	// second call returns the cached instance (covers the early-return branch)
	second := c.GetInstance(logWithCtx, "rep", map[InstanceName]interface{}{})
	assert.Equal(t, first, second)
}

func TestContainer_Release_SingletonNoOp(t *testing.T) {
	c := NewContainer(Config{InstanceType: Singleton})
	assert.NotPanics(t, func() {
		c.Release(logWithCtx, "anything", &userRep{})
	})
}

func TestContainer_Release_NotRegistered(t *testing.T) {
	c := NewContainer(Config{InstanceType: MultiInstance})
	assert.NotPanics(t, func() {
		c.Release(logWithCtx, "missing", &userRep{})
	})
}

func TestContainer_Release_WrongType(t *testing.T) {
	c := NewContainer(Config{InstanceType: MultiInstance})
	c.RegisterMultiInstance(logWithCtx, "rep", &sync.Pool{New: func() interface{} { return &userRep{} }})
	assert.NotPanics(t, func() {
		c.Release(logWithCtx, "rep", &orderRepo{})
	})
}

func TestDiSelfCheck_AutoWireFalse_Warning(t *testing.T) {
	// the WithCanSet+autowire:false path only logs a warning and continues.
	type fixture struct {
		Field *userRep `container:"autowire:false;resource:rep"`
	}
	c := NewContainer(Config{InstanceType: MultiInstance})
	c.RegisterMultiInstance(logWithCtx, "rep", &sync.Pool{New: func() interface{} { return &userRep{} }})
	c.RegisterMultiInstance(logWithCtx, "fix", &sync.Pool{New: func() interface{} { return &fixture{} }})
	assert.NoError(t, c.DiSelfCheck(logWithCtx, "fix"))
}

func TestDiAllFields_AutoWireFalse_Continue(t *testing.T) {
	type fixture struct {
		Field *userRep `container:"autowire:false;resource:rep"`
	}
	c := NewContainer(Config{InstanceType: MultiInstance})
	c.RegisterMultiInstance(logWithCtx, "rep", &sync.Pool{New: func() interface{} { return &userRep{} }})
	f := &fixture{}
	c.DiAllFields(logWithCtx, f, map[InstanceName]interface{}{})
	assert.Nil(t, f.Field)
}

func TestGetBoolTagFromContainer_NoTag(t *testing.T) {
	type fixture struct {
		Field string
	}
	v, exist := getBoolTagFromContainer(&fixture{}, 0, _CONTAINER, _AUTO_WIRE)
	assert.False(t, v)
	assert.False(t, exist)
}
