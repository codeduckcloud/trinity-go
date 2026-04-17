package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaginationDTO(t *testing.T) {
	p := NewPaginationDTO(20, 3, 120)
	assert.Equal(t, int64(120), p.Total)
	assert.Equal(t, 3, p.Current)
	assert.Equal(t, 6, p.TotalPage)
	assert.Equal(t, 20, p.PageSize)

	p2 := NewPaginationDTO(20, 1, 121)
	assert.Equal(t, 7, p2.TotalPage)
}
