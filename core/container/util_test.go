package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getTagByName(t *testing.T) {
	type args struct {
		object interface{}
		index  int
		name   Keyword
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "1",
			args: args{
				object: struct {
					Name string `test:"123"`
				}{},
				index: 0,
				name:  "test",
			},
			want:  "123",
			want1: true,
		},
		{
			name: "2",
			args: args{
				object: struct {
					Name  string `test:"123"`
					Name2 string `test:"12344ada"`
				}{},
				index: 1,
				name:  "test",
			},
			want:  "12344ada",
			want1: true,
		},
		{
			name: "2",
			args: args{
				object: &struct {
					Name  string `test:"123"`
					Name2 string `test:"12344ada"`
				}{},
				index: 1,
				name:  "test",
			},
			want:  "12344ada",
			want1: true,
		},
		{
			name: "3",
			args: args{
				object: &struct {
					Name  struct{ a string } `test:"123"`
					Name2 string             `test:"12344ada"`
				}{},
				index: 0,
				name:  "test",
			},
			want:  "123",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getTagByName(tt.args.object, tt.args.index, tt.args.name)
			if got != tt.want {
				t.Errorf("getTagByName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getTagByName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
	assert.Panics(t, func() { getTagByName("123", 1, "123") }, "panic")
}

func TestContainer_getResourceTag(t *testing.T) {
	type test3 struct {
		test1 string `container:"resource:1234"`
		test2 string `container:"resource:asdf"`
		test3 string `container:"resource:12fs"`
		test4 string `container:"resource:213"`
	}
	type args struct {
		obj   interface{}
		index int
	}
	tests := []struct {
		name   string
		fields *Container
		args   args
		want   string
	}{
		{
			name:   "1",
			fields: NewContainer(),
			args: args{
				obj:   &test3{},
				index: 0,
			},
			want: "1234",
		},
		{
			name:   "2",
			fields: NewContainer(),
			args: args{
				obj:   &test3{},
				index: 1,
			},
			want: "asdf",
		},
		{
			name:   "3",
			fields: NewContainer(),
			args: args{
				obj:   &test3{},
				index: 2,
			},
			want: "12fs",
		},
		{
			name:   "4",
			fields: NewContainer(),
			args: args{
				obj:   &test3{},
				index: 3,
			},
			want: "213",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := getStringTagFromContainerByKey(tt.args.obj, tt.args.index, _CONTAINER, _RESOURCE); got != tt.want {
				t.Errorf("Container.getAutoFreeTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeTag(t *testing.T) {
	type args struct {
		value string
		key   Keyword
	}
	tests := []struct {
		name     string
		args     args
		want     string
		wantBool bool
	}{
		{
			name: "1",
			args: args{
				value: "autowire:true",
				key:   _AUTO_WIRE,
			},
			want:     "true",
			wantBool: true,
		},
		{
			name: "2",
			args: args{
				value: "autowire:true;autowire:false;",
				key:   _AUTO_WIRE,
			},
			want:     "false",
			wantBool: true,
		},
		{
			name: "3",
			args: args{
				value: "autowire;",
				key:   _AUTO_WIRE,
			},
			want:     "",
			wantBool: true,
		},
		{
			name: "4",
			args: args{
				value: ":;",
				key:   _AUTO_WIRE,
			},
			want:     "",
			wantBool: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeTag(tt.args.value, tt.args.key)
			if got != tt.want {
				t.Errorf("decodeTag() value = %v, want %v", got, tt.want)
			}
			if ok != tt.wantBool {
				t.Errorf("decodeTag() isExist= %v, want %v", ok, tt.wantBool)
			}
		})
	}
}
