package utils

import (
	"testing"
)

func TestMD5Content(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				content: []byte("123"),
			},
			want: "ICy5YqxZB1uWSwcVLSNLcA==",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MD5Content(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("MD5Content() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MD5Content() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandUInt32(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateNonce()
			if (err != nil) != tt.wantErr {
				t.Errorf("RandUInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
