package e

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(1001, "bad request", "detail1", "detail2")
	assert.Equal(t, 1001, err.Code())
	assert.Equal(t, "bad request", err.Message())
	assert.Equal(t, []string{"detail1", "detail2"}, err.Details())
	assert.Equal(t, "error code: 1001, error msg: bad request, error details: [detail1 detail2]", err.Error())
}

func TestNewWithoutDetails(t *testing.T) {
	err := New(1002, "oops")
	assert.Equal(t, 1002, err.Code())
	assert.Equal(t, "oops", err.Message())
	assert.Empty(t, err.Details())
	assert.Contains(t, err.Error(), "oops")
}

func TestFromErr_Nil(t *testing.T) {
	assert.Nil(t, FromErr(nil))
}

func TestFromErr_StdError(t *testing.T) {
	wrapped := FromErr(errors.New("boom"))
	assert.NotNil(t, wrapped)
	assert.Equal(t, UnknownError, wrapped.Code())
	assert.Equal(t, "boom", wrapped.Message())
}

func TestFromErr_WrapError(t *testing.T) {
	original := New(4001, "already wrapped", "x")
	wrapped := FromErr(original)
	assert.Equal(t, original, wrapped)
}
