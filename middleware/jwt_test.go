package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJwt(t *testing.T) {
	assert.NotPanics(t, func() {
		Jwt()
	})
}
