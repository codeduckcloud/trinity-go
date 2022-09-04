package container

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_getAutoWireTag(t *testing.T) {
	type testAutoWire struct {
		test1 string `container:"autowire:true"`
		test2 string `container:"autowire:false"`
		test3 string `container:"autowire"`
		test4 string `container:""`
	}
	type args struct {
		obj   interface{}
		index int
	}
	tests := []struct {
		name   string
		fields *Container
		args   args
		want   bool
	}{
		{
			name:   "auto wire true",
			fields: NewContainer(),
			args: args{
				obj:   &testAutoWire{},
				index: 0,
			},
			want: true,
		},
		{
			name:   "auto wire false",
			fields: NewContainer(),
			args: args{
				obj:   &testAutoWire{},
				index: 1,
			},
			want: false,
		},
		{
			name:   "auto wire empty",
			fields: NewContainer(),
			args: args{
				obj:   &testAutoWire{},
				index: 2,
			},
			want: true,
		},
		{
			name:   "no auto wire tag",
			fields: NewContainer(),
			args: args{
				obj:   &testAutoWire{},
				index: 3,
			},
			want: true,
		},
		{
			name: "no auto wire tag",
			fields: NewContainer(Config{
				AutoWire: false,
			}),
			args: args{
				obj:   &testAutoWire{},
				index: 3,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.getAutoWireTag(tt.args.obj, tt.args.index); got != tt.want {
				t.Errorf("Container.getAutoWireTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

type userRep struct{}
type addressRepo struct{}
type orderRepo struct{}
type shipRepo struct{}

type userSrv struct {
	UserRep    *userRep    `container:"autowire:true;resource:userRep"`
	AddressSrv *addressSrv `container:"autowire:true;resource:addressSrv"`
	OrderSrv   *orderSrv   `container:"autowire:true;resource:orderSrv"`
}
type addressSrv struct {
	UserSrv     *userSrv     `container:"autowire:true;resource:userSrv"`
	AddressRepo *addressRepo `container:"autowire:true;resource:addressRepo"`
}
type shipSrv struct {
	UserSrv    *userSrv    `container:"autowire:true;resource:userSrv"`
	AddressSrv *addressSrv `container:"autowire:true;resource:addressSrv"`
	OrderSrv   *orderSrv   `container:"autowire:true;resource:orderSrv"`
	ShipRepo   *shipRepo   `container:"autowire:true;resource:shipRepo"`
}
type orderSrv struct {
	UserSrv    *userSrv    `container:"autowire:true;resource:userSrv"`
	AddressSrv *addressSrv `container:"autowire:true;resource:addressSrv"`
	ShipSrv    *shipSrv    `container:"autowire:true;resource:shipSrv"`
	OrderRepo  *orderRepo  `container:"autowire:true;resource:orderRepo"`
}

type userCtl struct {
	UserSrv *userSrv `container:"autowire:true;resource:userSrv"`
}
type addressCtl struct {
	AddressSrv *addressSrv `container:"autowire:true;resource:addressSrv"`
}
type orderCtl struct {
	OrderSrv *orderSrv `container:"autowire:true;resource:orderSrv"`
}
type shipCtl struct {
	ShipSrv *shipSrv `container:"autowire:true;resource:shipSrv"`
}

func TestContainer_GetInstance(t *testing.T) {
	{
		// with no pool registered
		s1 := &testShared1{}
		s2 := &testShared2{}
		s1.T = s2
		s2.T = s1
		s := NewContainer()
		injectMap := make(map[InstanceName]interface{})
		assert.Panics(t, func() { s.GetInstance(logWithCtx, "shared1", injectMap) })
	}
	{
		// with inject map
		s1 := &testShared1{}
		s2 := &testShared2{}
		s1.T = s2
		s2.T = s1
		s := NewContainer(Config{InstanceType: MultiInstance})
		s.RegisterMultiInstance(logWithCtx, "shared2", &sync.Pool{New: func() interface{} { return &testShared2{} }})
		injectMap := map[InstanceName]interface{}{
			"shared1": s1,
		}
		instance := s.GetInstance(logWithCtx, "shared1", injectMap)
		assert.Equal(t, s1, instance, "wrong instance ")
	}
	{
		// shared instance
		s1 := &testShared1{}
		s2 := &testShared2{}
		s1.T = s2
		s2.T = s1
		s := NewContainer(Config{InstanceType: MultiInstance})
		s.RegisterMultiInstance(logWithCtx, "shared1", &sync.Pool{New: func() interface{} { return &testShared1{} }})
		s.RegisterMultiInstance(logWithCtx, "shared2", &sync.Pool{New: func() interface{} { return &testShared2{} }})
		injectMap := make(map[InstanceName]interface{})
		instance := s.GetInstance(logWithCtx, "shared1", injectMap)
		assert.Equal(t, s1, instance, "wrong instance ")
		t1 := instance.(*testShared1)
		assert.Equal(t, fmt.Sprintf("%p", t1), fmt.Sprintf("%p", t1.T.T), "wrong ptr")
		assert.Equal(t, fmt.Sprintf("%p", t1.T), fmt.Sprintf("%p", t1.T.T.T), "wrong ptr")
		assert.NotPanics(t, func() { fmt.Printf("%p", t1.T.T.T.T.T.T.T.T.T.T.T.T.T.T.T.T.T.T.T) }, "inject failed")
	}
	{
		// ptr register
		type testRep struct{}
		type testSrv struct {
			TestRep *testRep `container:"autowire:true;resource:testRep"`
		}
		type testCtl struct {
			TestSrv *testSrv `container:"autowire:true;resource:testSrv"`
		}

		res := &testCtl{
			TestSrv: &testSrv{
				TestRep: &testRep{},
			},
		}
		s := NewContainer(Config{InstanceType: MultiInstance})
		s.RegisterMultiInstance(logWithCtx, "testRep", &sync.Pool{New: func() interface{} { return &testRep{} }})
		s.RegisterMultiInstance(logWithCtx, "testSrv", &sync.Pool{New: func() interface{} { return &testSrv{} }})
		s.RegisterMultiInstance(logWithCtx, "testCtl", &sync.Pool{New: func() interface{} { return &testCtl{} }})
		injectMap := make(map[InstanceName]interface{})
		instance := s.GetInstance(logWithCtx, "testCtl", injectMap)
		assert.Equal(t, res, instance, "wrong instance ")
		t1 := instance.(*testCtl)
		assert.NotPanics(t, func() { fmt.Printf("%p", t1.TestSrv.TestRep) }, "inject failed")
	}
	{
		// interface register
		type testRepI interface{}
		type testRep struct{}
		type testSrvI interface{}
		type testSrv struct {
			TestRep testRepI `container:"autowire:true;resource:testRepI"`
		}
		type testCtl struct {
			TestSrv testSrvI `container:"autowire:true;resource:testSrvI"`
		}

		res := &testCtl{
			TestSrv: &testSrv{
				TestRep: &testRep{},
			},
		}
		s := NewContainer(Config{InstanceType: MultiInstance})
		s.RegisterMultiInstance(logWithCtx, "testRepI", &sync.Pool{New: func() interface{} { return &testRep{} }})
		s.RegisterMultiInstance(logWithCtx, "testSrvI", &sync.Pool{New: func() interface{} { return &testSrv{} }})
		s.RegisterMultiInstance(logWithCtx, "testCtl", &sync.Pool{New: func() interface{} { return &testCtl{} }})
		injectMap := make(map[InstanceName]interface{})
		instance := s.GetInstance(logWithCtx, "testCtl", injectMap)
		assert.Equal(t, res, instance, "wrong instance ")
		t1 := instance.(*testCtl)
		assert.NotPanics(t, func() { fmt.Printf("%p", &t1.TestSrv) }, "inject failed")
		t2 := t1.TestSrv.(*testSrv)
		assert.NotPanics(t, func() { fmt.Printf("%p", &t2.TestRep) }, "inject failed")
	}

	{
		// blackbox
		uSrv := &userSrv{}
		aSrv := &addressSrv{}
		oSrv := &orderSrv{}
		sSrv := &shipSrv{}
		uSrv.AddressSrv = aSrv
		uSrv.OrderSrv = oSrv
		uSrv.UserRep = &userRep{}
		aSrv.UserSrv = uSrv
		aSrv.AddressRepo = &addressRepo{}
		oSrv.AddressSrv = aSrv
		oSrv.UserSrv = uSrv
		oSrv.ShipSrv = sSrv
		oSrv.OrderRepo = &orderRepo{}
		sSrv.AddressSrv = aSrv
		sSrv.ShipRepo = &shipRepo{}
		sSrv.OrderSrv = oSrv
		sSrv.UserSrv = uSrv
		userCtlRes := &userCtl{
			UserSrv: uSrv,
		}
		addressCtlRes := &addressCtl{
			AddressSrv: aSrv,
		}
		orderCtlRes := &orderCtl{
			OrderSrv: oSrv,
		}
		shipCtlRes := &shipCtl{
			ShipSrv: sSrv,
		}
		s := NewContainer(Config{InstanceType: MultiInstance})
		s.RegisterMultiInstance(logWithCtx, "userRep", &sync.Pool{New: func() interface{} { return &userRep{} }})
		s.RegisterMultiInstance(logWithCtx, "addressRepo", &sync.Pool{New: func() interface{} { return &addressRepo{} }})
		s.RegisterMultiInstance(logWithCtx, "orderRepo", &sync.Pool{New: func() interface{} { return &orderRepo{} }})
		s.RegisterMultiInstance(logWithCtx, "shipRepo", &sync.Pool{New: func() interface{} { return &shipRepo{} }})
		s.RegisterMultiInstance(logWithCtx, "userSrv", &sync.Pool{New: func() interface{} { return &userSrv{} }})
		s.RegisterMultiInstance(logWithCtx, "addressSrv", &sync.Pool{New: func() interface{} { return &addressSrv{} }})
		s.RegisterMultiInstance(logWithCtx, "orderSrv", &sync.Pool{New: func() interface{} { return &orderSrv{} }})
		s.RegisterMultiInstance(logWithCtx, "shipSrv", &sync.Pool{New: func() interface{} { return &shipSrv{} }})
		s.RegisterMultiInstance(logWithCtx, "userCtl", &sync.Pool{New: func() interface{} { return &userCtl{} }})
		s.RegisterMultiInstance(logWithCtx, "addressCtl", &sync.Pool{New: func() interface{} { return &addressCtl{} }})
		s.RegisterMultiInstance(logWithCtx, "orderCtl", &sync.Pool{New: func() interface{} { return &orderCtl{} }})
		s.RegisterMultiInstance(logWithCtx, "shipCtl", &sync.Pool{New: func() interface{} { return &shipCtl{} }})
		if err := s.InstanceDISelfCheck(logWithCtx); err != nil {
			assert.Error(t, err, "self check error ")
		}
		{
			injectMap := make(map[InstanceName]interface{})
			instance := s.GetInstance(logWithCtx, "userCtl", injectMap)
			assert.Equal(t, userCtlRes, instance, "wrong instance ")
			t1 := instance.(*userCtl)
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.UserSrv.UserRep) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.UserSrv.AddressSrv) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.UserSrv.OrderSrv) }, "inject failed")
			for k, v := range injectMap {
				s.Release(logWithCtx, k, v)
			}
			assert.Panics(t, func() { fmt.Printf("%p", t1.UserSrv.UserRep) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.UserSrv.AddressSrv) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.UserSrv.OrderSrv) }, "release failed")
		}
		{
			injectMap := make(map[InstanceName]interface{})
			instance := s.GetInstance(logWithCtx, "addressCtl", injectMap)
			assert.Equal(t, addressCtlRes, instance, "wrong instance ")
			t1 := instance.(*addressCtl)
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.AddressSrv.AddressRepo) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.AddressSrv.UserSrv) }, "inject failed")
			for k, v := range injectMap {
				s.Release(logWithCtx, k, v)
			}
			assert.Panics(t, func() { fmt.Printf("%p", t1.AddressSrv.AddressRepo) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.AddressSrv.UserSrv) }, "release failed")
		}
		{
			injectMap := make(map[InstanceName]interface{})
			instance := s.GetInstance(logWithCtx, "orderCtl", injectMap)
			assert.Equal(t, orderCtlRes, instance, "wrong instance ")
			t1 := instance.(*orderCtl)
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.OrderSrv.UserSrv) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.OrderSrv.AddressSrv) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.OrderSrv.ShipSrv) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.OrderSrv.OrderRepo) }, "inject failed")
			for k, v := range injectMap {
				s.Release(logWithCtx, k, v)
			}
			assert.Panics(t, func() { fmt.Printf("%p", t1.OrderSrv.UserSrv) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.OrderSrv.AddressSrv) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.OrderSrv.ShipSrv) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.OrderSrv.OrderRepo) }, "release failed")
		}
		{
			injectMap := make(map[InstanceName]interface{})
			instance := s.GetInstance(logWithCtx, "shipCtl", injectMap)
			assert.Equal(t, shipCtlRes, instance, "wrong instance ")
			t1 := instance.(*shipCtl)
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.ShipSrv.UserSrv) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.ShipSrv.AddressSrv) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.ShipSrv.ShipRepo) }, "inject failed")
			assert.NotPanics(t, func() { fmt.Printf("%p", t1.ShipSrv.OrderSrv) }, "inject failed")
			for k, v := range injectMap {
				s.Release(logWithCtx, k, v)
			}
			assert.Panics(t, func() { fmt.Printf("%p", t1.ShipSrv.UserSrv) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.ShipSrv.AddressSrv) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.ShipSrv.ShipRepo) }, "release failed")
			assert.Panics(t, func() { fmt.Printf("%p", t1.ShipSrv.OrderSrv) }, "release failed")
		}

	}
}

func TestContainer_CheckInstanceNameIfExist(t *testing.T) {
	type fields struct {
		c           *Config
		poolMap     map[InstanceName]*sync.Pool
		poolTypeMap map[InstanceName]reflect.Type
	}
	type args struct {
		instanceName InstanceName
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "not exist",
			fields: fields{
				c:           _DefaultConfig,
				poolMap:     make(map[InstanceName]*sync.Pool),
				poolTypeMap: make(map[InstanceName]reflect.Type),
			},
			args: args{
				instanceName: InstanceName("test"),
			},
			want: false,
		},
		{
			name: " exist",
			fields: fields{
				c: _DefaultConfig,
				poolMap: map[InstanceName]*sync.Pool{
					InstanceName("test"): nil,
				},
				poolTypeMap: make(map[InstanceName]reflect.Type),
			},
			args: args{
				instanceName: InstanceName("test"),
			},
			want: true,
		},
		{
			name: " exist",
			fields: fields{
				c: _DefaultConfig,
				poolMap: map[InstanceName]*sync.Pool{
					InstanceName("test"): &sync.Pool{},
				},
				poolTypeMap: make(map[InstanceName]reflect.Type),
			},
			args: args{
				instanceName: InstanceName("test"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Container{
				c:           tt.fields.c,
				poolMap:     tt.fields.poolMap,
				poolTypeMap: tt.fields.poolTypeMap,
			}
			if got := s.CheckInstanceNameIfExist(tt.args.instanceName); got != tt.want {
				t.Errorf("Container.CheckInstanceNameIfExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainer_RegisterInstance(t *testing.T) {
	type fields struct {
		c                      *Config
		poolMap                map[InstanceName]*sync.Pool
		poolTypeMap            map[InstanceName]reflect.Type
		instanceMap            map[InstanceName]interface{}
		instanceMapInitialized map[InstanceName]interface{}
	}
	type args struct {
		ctx          context.Context
		instanceName InstanceName
		instance     interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "singleton conflict",
			fields: fields{
				c:                      _DefaultConfig,
				poolMap:                make(map[InstanceName]*sync.Pool),
				poolTypeMap:            make(map[InstanceName]reflect.Type),
				instanceMap:            make(map[InstanceName]interface{}),
				instanceMapInitialized: make(map[InstanceName]interface{}),
			},
			args: args{
				ctx:          logWithCtx,
				instanceName: InstanceName("123"),
				instance:     &userRep{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Container{
				c:                      tt.fields.c,
				poolMap:                tt.fields.poolMap,
				poolTypeMap:            tt.fields.poolTypeMap,
				instanceMap:            tt.fields.instanceMap,
				instanceMapInitialized: tt.fields.instanceMapInitialized,
			}
			s.RegisterInstance(tt.args.ctx, tt.args.instanceName, tt.args.instance)
			srv := s.GetInstance(tt.args.ctx, tt.args.instanceName, map[InstanceName]interface{}{})
			assert.Equal(t, &userRep{}, srv, "register instance error ")
		})
	}
}
