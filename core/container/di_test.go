package container

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type inject1 struct {
	name string
}
type injectInterface interface {
	A()
	B()
	C()
}
type injectInterfaceImpl struct{}

func (i injectInterfaceImpl) A() {}
func (i injectInterfaceImpl) B() {}
func (i injectInterfaceImpl) C() {}

type testInjectErr1 struct {
	test1 inject1 `container:"autowire:true"`
}
type testInjectErr2 struct {
	test1 inject1 `container:"autowire:true;resource:inject1"`
}
type testInjectErr3 struct {
	Test1 inject1 `container:"autowire:true;resource:inject1"`
}
type testInject3 struct {
	Test1 inject1 `container:"autowire:true;resource:inject1"`
}
type testInject4 struct {
	Test1 *inject1 `container:"autowire:true;resource:inject1"`
}

type testInject5 struct {
	Test1 injectInterface `container:"autowire:true;resource:inject1"`
}

type testInject6 struct {
	Test1 interface{} `container:"autowire:true;resource:inject1"`
}

type testShared1 struct {
	T *testShared2 `container:"autowire:true;resource:shared2"`
}

type testShared2 struct {
	T *testShared1 `container:"autowire:true;resource:shared1"`
}

func TestContainer_DiFree(t *testing.T) {
	a := inject1{"1"}
	empty := inject1{}
	{
		obj := struct{}{}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Equal(t, struct{}{}, obj, "test case: di free empty struct")
	}
	{
		obj := struct {
			test1 inject1 `container:"autowire:true"`
		}{a}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Equal(t, inject1{"1"}, obj.test1, "test case: di free private struct")
	}
	{
		obj := struct {
			Test1 inject1 `container:"autowire:true"`
		}{a}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Equal(t, empty, obj.Test1, "test case: di free success struct")
	}
	{
		obj := struct {
			Test1 inject1 `container:"autowire:false"`
		}{a}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Equal(t, a, obj.Test1, "test case: di free skip auto wire false")
	}
	{
		obj := struct {
			Test1 *inject1 `container:"autowire:true"`
		}{&a}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Nil(t, obj.Test1, "test case: di free ptr success")
	}
	{
		obj := struct {
			Test1 string `container:"autowire:true"`
		}{"1"}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Equal(t, "", obj.Test1, "test case: di free string success")
	}
	{
		s := "1"
		obj := struct {
			Test1 *string `container:"autowire:true"`
		}{&s}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Nil(t, obj.Test1, "test case: di free string ptr success")
	}
	{
		obj := struct {
			T string
		}{"1"}
		NewContainer().DiFree(logWithCtx, &obj)
		assert.Equal(t, "1", obj.T, "test case: di free no container tag , skipped cases")
	}
}

type SingletonSelfCheck4 interface {
	Get() string
}
type SingletonSelfCheck1 struct {
	A *SingletonSelfCheck2 `container:"autowire:true;resource:SingletonSelfCheck2"`
	B *SingletonSelfCheck3 `container:"autowire:true;resource:SingletonSelfCheck3"`
	C SingletonSelfCheck4  `container:"autowire:true;resource:SingletonSelfCheck4"`
	D singletonSelfCheck5
}
type SingletonSelfCheck2 struct{}
type SingletonSelfCheck3 struct{}

type singletonSelfCheck4 struct{}
type singletonSelfCheck5 struct{}

func (s *singletonSelfCheck4) Get() string {
	return ""
}

func TestContainer_DiSelfCheck(t *testing.T) {
	type fields struct {
		c         *Container
		instances map[InstanceName]*sync.Pool
	}
	type args struct {
		instanceName InstanceName
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "instance not in map ",
			fields: fields{
				c:         NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "instance `instance1` not exist in pool map",
		},
		{
			name: "instance cannot be addressable",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} { return testInjectErr1{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "the object to be injected container.testInjectErr1 should be addressable",
		},
		{
			name: "resource tag not set ",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} { return &testInjectErr1{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: instance1 index: 0 objectName: *container.testInjectErr1.test1.(container.inject1), the resource tag not exist in container",
		},
		{
			name: "resource name not register ",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} { return &testInjectErr3{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: instance1 index: 0 objectName: *container.testInjectErr3.Test1.(container.inject1), resource name: inject1 not register in container ",
		},
		{
			name: "private param",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} { return &testInjectErr2{} },
					},
					"inject1": {
						New: func() interface{} { return &testInjectErr2{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: instance1 index: 0 objectName: *container.testInjectErr2.test1.(container.inject1), private param",
		},
		{
			name: "object is not null ",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} {
							return &testInject3{
								Test1: inject1{
									name: "1",
								},
							}
						},
					},
					"inject1": {
						New: func() interface{} { return &testInjectErr2{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: instance1 index: 0 objectName: *container.testInject3.Test1.(container.inject1), the param to be injected is not null",
		},
		{
			name: "struct inject type not equal",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} { return &testInject3{} },
					},
					"inject1": {
						New: func() interface{} { return &testInject3{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: instance1 index: 0 objectName: *container.testInject3.Test1.(container.inject1), resource name: inject1 type not same, expected: container.inject1 actual: *container.testInject3",
		},
		{
			name: "ptr inject type not equal",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} { return &testInject4{} },
					},
					"inject1": {
						New: func() interface{} { return &testInject3{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: instance1 index: 0 objectName: *container.testInject4.Test1.(*container.inject1), resource name: inject1 type not same, expected: *container.inject1 actual: *container.testInject3",
		},
		{
			name: "interface inject type not implement",
			fields: fields{
				c: NewContainer(Config{InstanceType: MultiInstance}),
				instances: map[InstanceName]*sync.Pool{
					"instance1": {
						New: func() interface{} { return &testInject5{} },
					},
					"inject1": {
						New: func() interface{} { return &testInject3{} },
					},
				},
			},
			args: args{
				instanceName: "instance1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: instance1 index: 0 objectName: *container.testInject5.Test1.(container.injectInterface), resource name: inject1 type:  not implement the interface injectInterface",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.fields.instances {
				tt.fields.c.RegisterMultiInstance(logWithCtx, k, v)
			}
			if err := tt.fields.c.DiSelfCheck(logWithCtx, tt.args.instanceName); err != nil {
				if tt.wantErr {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				} else {
					t.Error("unexpected error ")
					t.FailNow()
				}
			}

		})
	}
}

func TestContainer_Singleton_DiSelfCheck(t *testing.T) {
	type fields struct {
		c         *Container
		instances map[InstanceName]interface{}
	}
	type args struct {
		instanceName InstanceName
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "singleton",
			fields: fields{
				c: NewContainer(Config{InstanceType: Singleton}),
				instances: map[InstanceName]interface{}{
					"SingletonSelfCheck1": &SingletonSelfCheck1{},
					"SingletonSelfCheck2": &SingletonSelfCheck2{},
					"SingletonSelfCheck3": &SingletonSelfCheck3{},
					"SingletonSelfCheck4": &singletonSelfCheck4{},
				},
			},
			args: args{
				instanceName: "SingletonSelfCheck1",
			},
			wantErr: false,
		},
		{
			name: "type not equal",
			fields: fields{
				c: NewContainer(Config{InstanceType: Singleton}),
				instances: map[InstanceName]interface{}{
					"SingletonSelfCheck1": &SingletonSelfCheck1{},
					"SingletonSelfCheck2": SingletonSelfCheck2{},
					"SingletonSelfCheck3": &SingletonSelfCheck3{},
					"SingletonSelfCheck4": &singletonSelfCheck4{},
				},
			},
			args: args{
				instanceName: "SingletonSelfCheck1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: SingletonSelfCheck1 index: 0 objectName: *container.SingletonSelfCheck1.A.(*container.SingletonSelfCheck2), resource name: SingletonSelfCheck2 type not same, expected: *container.SingletonSelfCheck2 actual: container.SingletonSelfCheck2",
		},
		{
			name: "interface not implement ",
			fields: fields{
				c: NewContainer(Config{InstanceType: Singleton}),
				instances: map[InstanceName]interface{}{
					"SingletonSelfCheck1": &SingletonSelfCheck1{},
					"SingletonSelfCheck2": &SingletonSelfCheck2{},
					"SingletonSelfCheck3": &SingletonSelfCheck3{},
					"SingletonSelfCheck4": singletonSelfCheck4{},
				},
			},
			args: args{
				instanceName: "SingletonSelfCheck1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: SingletonSelfCheck1 index: 2 objectName: *container.SingletonSelfCheck1.C.(container.SingletonSelfCheck4), resource name: SingletonSelfCheck4 type: singletonSelfCheck4 not implement the interface SingletonSelfCheck4",
		},
		{
			name: "missing instance ",
			fields: fields{
				c: NewContainer(Config{InstanceType: Singleton}),
				instances: map[InstanceName]interface{}{
					"SingletonSelfCheck1": &SingletonSelfCheck1{},
					"SingletonSelfCheck2": &SingletonSelfCheck2{},
					"SingletonSelfCheck4": singletonSelfCheck4{},
				},
			},
			args: args{
				instanceName: "SingletonSelfCheck1",
			},
			wantErr:    true,
			wantErrMsg: "self check error: instanceName: SingletonSelfCheck1 index: 1 objectName: *container.SingletonSelfCheck1.B.(*container.SingletonSelfCheck3), resource name: SingletonSelfCheck3 not register in container ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.fields.instances {
				tt.fields.c.RegisterInstance(logWithCtx, k, v)
			}
			if err := tt.fields.c.DiSelfCheck(logWithCtx, tt.args.instanceName); err != nil {
				if tt.wantErr {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				} else {
					t.Error("unexpected error ,err", err.Error())
					t.FailNow()
				}
			}
		})
	}
}

func TestContainer_DiAllFields(t *testing.T) {
	p1 := &inject1{}
	s1 := &testShared1{}
	s2 := &testShared2{}
	s1.T = s2
	s2.T = s1
	type fields struct {
		c       *Config
		poolMap map[InstanceName]*sync.Pool
	}
	type args struct {
		dest         interface{}
		injectingMap map[InstanceName]interface{}
	}
	type want struct {
		instance  interface{}
		injectMap map[InstanceName]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "di ptr",
			fields: fields{
				c: &Config{
					AutoWire:        true,
					JsonTagKeyword:  _CONTAINER,
					AutoWireKeyword: _AUTO_WIRE,
					ResourceKeyword: _RESOURCE,
					InstanceType:    MultiInstance,
				},
				poolMap: map[InstanceName]*sync.Pool{
					"inject1": {
						New: func() interface{} { return &inject1{} },
					},
				},
			},
			args: args{
				dest:         &testInject4{},
				injectingMap: make(map[InstanceName]interface{}),
			},
			want: want{
				instance: &testInject4{
					Test1: &inject1{},
				},
				injectMap: map[InstanceName]interface{}{
					"inject1": &inject1{},
				},
			},
		},
		{
			name: "di with inject map instance",
			fields: fields{
				c: &Config{
					AutoWire:        true,
					JsonTagKeyword:  _CONTAINER,
					AutoWireKeyword: _AUTO_WIRE,
					ResourceKeyword: _RESOURCE,
					InstanceType:    MultiInstance,
				},
				poolMap: map[InstanceName]*sync.Pool{
					"inject1": {
						New: func() interface{} { return &inject1{} },
					},
				},
			},
			args: args{
				dest: &testInject4{},
				injectingMap: map[InstanceName]interface{}{
					"inject1": p1,
				},
			},
			want: want{
				instance: &testInject4{
					Test1: p1,
				},
				injectMap: map[InstanceName]interface{}{
					"inject1": p1,
				},
			},
		},
		{
			name: "di with inject map instance",
			fields: fields{
				c: &Config{
					AutoWire:        true,
					JsonTagKeyword:  _CONTAINER,
					AutoWireKeyword: _AUTO_WIRE,
					ResourceKeyword: _RESOURCE,
					InstanceType:    MultiInstance,
				},
				poolMap: map[InstanceName]*sync.Pool{
					"shared1": {
						New: func() interface{} { return &testShared1{} },
					},
					"shared2": {
						New: func() interface{} { return &testShared2{} },
					},
				},
			},
			args: args{
				dest: &testShared1{},
				injectingMap: map[InstanceName]interface{}{
					"shared1": s1,
				},
			},
			want: want{
				instance: s1,
				injectMap: map[InstanceName]interface{}{
					"shared1": s1,
					"shared2": s2,
				},
			},
		},
		{
			name: "di with interface",
			fields: fields{
				c: &Config{
					AutoWire:        true,
					JsonTagKeyword:  _CONTAINER,
					AutoWireKeyword: _AUTO_WIRE,
					ResourceKeyword: _RESOURCE,
					InstanceType:    MultiInstance,
				},
				poolMap: map[InstanceName]*sync.Pool{
					"shared1": {
						New: func() interface{} { return &testShared1{} },
					},
					"shared2": {
						New: func() interface{} { return &testShared2{} },
					},
					"inject1": {
						New: func() interface{} {
							return inject1{
								name: "1",
							}
						},
					},
				},
			},
			args: args{
				dest:         &testInject3{},
				injectingMap: map[InstanceName]interface{}{},
			},
			want: want{
				instance: &testInject3{
					Test1: inject1{
						name: "1",
					},
				},
				injectMap: map[InstanceName]interface{}{
					"inject1": inject1{
						name: "1",
					},
				},
			},
		},
		{
			name: "di with interface",
			fields: fields{
				c: &Config{
					AutoWire:        true,
					JsonTagKeyword:  _CONTAINER,
					AutoWireKeyword: _AUTO_WIRE,
					ResourceKeyword: _RESOURCE,
					InstanceType:    MultiInstance,
				},
				poolMap: map[InstanceName]*sync.Pool{
					"shared1": {
						New: func() interface{} { return &testShared1{} },
					},
					"shared2": {
						New: func() interface{} { return &testShared2{} },
					},
					"inject1": {
						New: func() interface{} {
							return &inject1{
								name: "1",
							}
						},
					},
				},
			},
			args: args{
				dest:         &testInject4{},
				injectingMap: map[InstanceName]interface{}{},
			},
			want: want{
				instance: &testInject4{
					Test1: &inject1{
						name: "1",
					},
				},
				injectMap: map[InstanceName]interface{}{
					"inject1": &inject1{
						name: "1",
					},
				},
			},
		},
		{
			name: "di with empty interface",
			fields: fields{
				c: &Config{
					AutoWire:        true,
					JsonTagKeyword:  _CONTAINER,
					AutoWireKeyword: _AUTO_WIRE,
					ResourceKeyword: _RESOURCE,
					InstanceType:    MultiInstance,
				},
				poolMap: map[InstanceName]*sync.Pool{
					"shared1": {
						New: func() interface{} { return &testShared1{} },
					},
					"shared2": {
						New: func() interface{} { return &testShared2{} },
					},
					"inject1": {
						New: func() interface{} {
							return &injectInterfaceImpl{}
						},
					},
				},
			},
			args: args{
				dest:         &testInject6{},
				injectingMap: map[InstanceName]interface{}{},
			},
			want: want{
				instance: &testInject6{
					Test1: &injectInterfaceImpl{},
				},
				injectMap: map[InstanceName]interface{}{
					"inject1": &injectInterfaceImpl{},
				},
			},
		},
		{
			name: "di with empty interface 2 ",
			fields: fields{
				c: &Config{
					AutoWire:        true,
					JsonTagKeyword:  _CONTAINER,
					AutoWireKeyword: _AUTO_WIRE,
					ResourceKeyword: _RESOURCE,
					InstanceType:    MultiInstance,
				},
				poolMap: map[InstanceName]*sync.Pool{
					"shared1": {
						New: func() interface{} { return &testShared1{} },
					},
					"shared2": {
						New: func() interface{} { return &testShared2{} },
					},
					"inject1": {
						New: func() interface{} {
							return inject1{}
						},
					},
				},
			},
			args: args{
				dest:         &testInject6{},
				injectingMap: map[InstanceName]interface{}{},
			},
			want: want{
				instance: &testInject6{
					Test1: inject1{},
				},
				injectMap: map[InstanceName]interface{}{
					"inject1": inject1{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Container{
				c:       tt.fields.c,
				poolMap: tt.fields.poolMap,
			}
			s.DiAllFields(logWithCtx, tt.args.dest, tt.args.injectingMap)
			assert.Equal(t, tt.want.instance, tt.args.dest, "wrong instance")
			assert.Equal(t, tt.want.injectMap, tt.args.injectingMap, "wrong inject map ")
		})
	}
}

func TestContainer_Singleton_DiAllFields(t *testing.T) {
	type fields struct {
		c         *Container
		instances map[InstanceName]interface{}
	}
	type args struct {
		dest interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "singleton",
			fields: fields{
				c: NewContainer(Config{InstanceType: Singleton}),
				instances: map[InstanceName]interface{}{
					"SingletonSelfCheck1": &SingletonSelfCheck1{},
					"SingletonSelfCheck2": &SingletonSelfCheck2{},
					"SingletonSelfCheck3": &SingletonSelfCheck3{},
					"SingletonSelfCheck4": &singletonSelfCheck4{},
				},
			},
			args: args{
				dest: &SingletonSelfCheck1{},
			},
			want: &SingletonSelfCheck1{
				A: &SingletonSelfCheck2{},
				B: &SingletonSelfCheck3{},
				C: &singletonSelfCheck4{},
				D: singletonSelfCheck5{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.fields.instances {
				tt.fields.c.RegisterInstance(logWithCtx, k, v)
			}
			tt.fields.c.DiAllFields(logWithCtx, tt.args.dest, make(map[InstanceName]interface{}))
			assert.Equal(t, tt.want, tt.args.dest, "wrong injected")
		})
	}
}
