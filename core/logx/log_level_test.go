package logx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{name: "panic", input: "panic", want: PanicLevel},
		{name: "fatal", input: "fatal", want: FatalLevel},
		{name: "error", input: "ERROR", want: ErrorLevel},
		{name: "warn", input: "warn", want: WarnLevel},
		{name: "warning", input: "Warning", want: WarnLevel},
		{name: "info", input: "info", want: InfoLevel},
		{name: "debug", input: "debug", want: DebugLevel},
		{name: "trace", input: "trace", want: TraceLevel},
		{name: "invalid", input: "test", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestLevel_UnmarshalText(t *testing.T) {
	var lvl Level
	assert.NoError(t, lvl.UnmarshalText([]byte("debug")))
	assert.Equal(t, DebugLevel, lvl)

	assert.Error(t, lvl.UnmarshalText([]byte("unknown-level")))
}

func TestLevel_MarshalText(t *testing.T) {
	tests := []struct {
		level   Level
		want    string
		wantErr bool
	}{
		{TraceLevel, "trace", false},
		{DebugLevel, "debug", false},
		{InfoLevel, "info", false},
		{WarnLevel, "warning", false},
		{ErrorLevel, "error", false},
		{FatalLevel, "fatal", false},
		{PanicLevel, "panic", false},
		{Level(99), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, err := tt.level.MarshalText()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}
