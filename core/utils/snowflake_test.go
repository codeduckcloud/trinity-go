package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSnowflakeID(t *testing.T) {
	id1 := GetSnowflakeID()
	id2 := GetSnowflakeID()
	assert.NotZero(t, id1)
	assert.NotZero(t, id2)
}
