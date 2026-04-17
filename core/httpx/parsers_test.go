package httpx

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type customHeaderParser struct {
	called bool
}

func (c *customHeaderParser) Exist(h http.Header, k string) bool { c.called = true; return false }
func (c *customHeaderParser) Get(h http.Header, k string) string { return "" }

type customQueryParser struct {
	called bool
}

func (c *customQueryParser) Exist(h url.Values, k string) bool { c.called = true; return false }
func (c *customQueryParser) Get(h url.Values, k string) string { return "" }

func TestSetHeaderParser(t *testing.T) {
	original := _defaultHeaderParser
	defer SetHeaderParser(original)

	custom := &customHeaderParser{}
	SetHeaderParser(custom)
	assert.Equal(t, custom, _defaultHeaderParser)
}

func TestSetQueryParser(t *testing.T) {
	original := _defaultQueryParser
	defer SetQueryParser(original)

	custom := &customQueryParser{}
	SetQueryParser(custom)
	assert.Equal(t, custom, _defaultQueryParser)
}

func TestNewWriter(t *testing.T) {
	w := NewWriter()
	assert.NotNil(t, w)
	assert.Nil(t, w.Header())
	n, err := w.Write([]byte("hi"))
	assert.Equal(t, 0, n)
	assert.NoError(t, err)
	w.WriteHeader(200)
}
