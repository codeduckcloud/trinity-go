package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMustXMLMarshal(t *testing.T) {
	type T struct {
		A string `xml:"abc"`
	}
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantPanic bool
	}{
		{
			name: "1",
			args: args{
				obj: struct{ abc string }{abc: "123"},
			},
			wantPanic: true,
		},
		{
			name: "1",
			args: args{
				obj: &T{A: "123"},
			},
			want: "<T><abc>123</abc></T>",
		},
		{
			name: "1",
			args: args{
				obj: T{A: "123"},
			},
			want: "<T><abc>123</abc></T>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { MustXMLMarshal(tt.args.obj) }, "wrong ")
			} else {
				got := MustXMLMarshal(tt.args.obj)
				assert.Equal(t, tt.want, got, "wrong ")
			}
		})
	}
}

func TestMustJSONMarshal(t *testing.T) {
	type T struct {
		A string `json:"abc"`
	}
	type test struct {
		A string `json:"abc"`
	}
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantPanic bool
	}{
		{
			name: "1",
			args: args{
				obj: make(chan int),
			},
			wantPanic: true,
		},
		{
			name: "1",
			args: args{
				obj: &T{A: "123"},
			},
			want: "{\"abc\":\"123\"}",
		},
		{
			name: "1",
			args: args{
				obj: T{A: "123"},
			},
			want: "{\"abc\":\"123\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { MustJSONMarshal(tt.args.obj) }, "wrong ")
			} else {
				got := MustJSONMarshal(tt.args.obj)
				assert.Equal(t, tt.want, got, "wrong ")
			}
		})
	}
}

func TestMustFormatTime(t *testing.T) {
	tTime, _ := time.Parse("2006", "2012")
	type args struct {
		layout     string
		timeString string
	}
	tests := []struct {
		name      string
		args      args
		want      *time.Time
		wantPanic bool
	}{
		// TODO: Add test cases.
		{
			name:      "1",
			args:      args{layout: "123", timeString: "123"},
			wantPanic: true,
		},
		{
			name: "1",
			args: args{layout: "2006", timeString: "2012"},
			want: &tTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { MustFormatTime(tt.args.layout, tt.args.timeString) }, "wrong ")
			} else {
				got := MustFormatTime(tt.args.layout, tt.args.timeString)
				assert.Equal(t, tt.want, got, "wrong ")
			}
		})
	}
}

func TestMustParseBool(t *testing.T) {
	type args struct {
		boolString string
	}
	tests := []struct {
		name      string
		args      args
		want      bool
		wantPanic bool
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				boolString: "1",
			},
			want: true,
		},
		{
			name: "1",
			args: args{
				boolString: "false",
			},
			want: false,
		},
		{
			name: "1",
			args: args{
				boolString: "123124",
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { MustParseBool(tt.args.boolString) }, "wrong ")
			} else {
				got := MustParseBool(tt.args.boolString)
				assert.Equal(t, tt.want, got, "wrong ")
			}
		})
	}
}

func TestMustParseFloat64(t *testing.T) {
	type args struct {
		floatString string
	}
	tests := []struct {
		name      string
		args      args
		want      float64
		wantPanic bool
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				floatString: "123.123",
			},
			want: 123.123,
		},
		{
			name: "2",
			args: args{
				floatString: "abs",
			},
			wantPanic: true,
		},
		{
			name: "3",
			args: args{
				floatString: "",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { MustParseFloat64(tt.args.floatString) }, "wrong ")
			} else {
				got := MustParseFloat64(tt.args.floatString)
				assert.Equal(t, tt.want, got, "wrong ")
			}
		})
	}
}
