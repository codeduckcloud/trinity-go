package logx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionLogger(t *testing.T) {
	logger := NewLogrusLogger()
	ctx := NewCtx(logger)

	var gotLogger Logger
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLogger = FromCtx(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mw := SessionLogger(ctx)(next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, logger, gotLogger)
}
