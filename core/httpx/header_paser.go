package httpx

import (
	"net/http"
	"net/textproto"
)

var (
	_defaultHeaderParser HeaderParser = &DefaultHeaderParser{}
)

func SetHeaderParser(h HeaderParser) {
	_defaultHeaderParser = h
}

type HeaderParser interface {
	Exist(h http.Header, k string) bool
	Get(h http.Header, k string) string
}

type DefaultHeaderParser struct {
}

func (p *DefaultHeaderParser) Exist(h http.Header, k string) bool {
	_, exist := h[textproto.CanonicalMIMEHeaderKey(k)]
	return exist
}

func (p *DefaultHeaderParser) Get(h http.Header, k string) string {
	return h.Get(k)
}
