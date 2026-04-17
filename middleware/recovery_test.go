package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codeduckcloud/trinity-go/core/httpx"
	"github.com/stretchr/testify/assert"
)

func TestRecovery_NoPanic(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	mw := Recovery()(next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTeapot, rr.Code)
}

func TestRecovery_Panic(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	mw := Recovery()(next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)
	assert.Equal(t, httpx.DefaultHttpErrorCode, rr.Code)
	assert.Contains(t, rr.Body.String(), "panic")
}
