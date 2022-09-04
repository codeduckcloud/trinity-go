package httpx

import (
	"net/url"
)

var (
	_defaultQueryParser QueryParser = &DefaultQueryParser{}
)

func SetQueryParser(h QueryParser) {
	_defaultQueryParser = h
}

type QueryParser interface {
	Exist(h url.Values, k string) bool
	Get(h url.Values, k string) string
}

type DefaultQueryParser struct {
}

func (p *DefaultQueryParser) Exist(h url.Values, k string) bool {
	_, exist := h[k]
	return exist
}

func (p *DefaultQueryParser) Get(h url.Values, k string) string {
	return h.Get(k)
}
