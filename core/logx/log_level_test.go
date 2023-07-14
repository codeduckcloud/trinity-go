package logx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLevel(t *testing.T) {
	got, err := ParseLevel("test")
	assert.Equal(t, "", got)
	assert.Equal(t, nil, err)
}
