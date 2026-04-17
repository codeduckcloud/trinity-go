package httpx

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/codeduckcloud/trinity-go/core/e"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
)

func TestJsonResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	JsonResponse(rr, 201, map[string]interface{}{"a": 1})
	assert.Equal(t, 201, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, float64(1), out["a"])
}

func TestJsonResponse_Panics(t *testing.T) {
	rr := httptest.NewRecorder()
	assert.Panics(t, func() {
		JsonResponse(rr, 200, make(chan int))
	})
}

func TestXMLResponse(t *testing.T) {
	type payload struct {
		Name string `xml:"name"`
	}
	rr := httptest.NewRecorder()
	XMLResponse(rr, 202, payload{Name: "hi"})
	assert.Equal(t, 202, rr.Code)
	assert.Equal(t, "application/xml", rr.Header().Get("Content-Type"))

	var p payload
	assert.NoError(t, xml.Unmarshal(rr.Body.Bytes(), &p))
	assert.Equal(t, "hi", p.Name)
}

func TestXMLResponse_Panics(t *testing.T) {
	rr := httptest.NewRecorder()
	assert.Panics(t, func() {
		XMLResponse(rr, 200, make(chan int))
	})
}

func TestHttpResponseErr(t *testing.T) {
	rr := httptest.NewRecorder()
	HttpResponseErr(context.Background(), rr, errors.New("boom"))
	assert.Equal(t, DefaultHttpErrorCode, rr.Code)

	var resp Response
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Error)
	assert.Equal(t, e.UnknownError, resp.Error.Code)
	assert.Equal(t, "boom", resp.Error.Message)
}

func TestHttpResponseErr_WithCustomStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), HttpxContext, &Context{code: 418})
	HttpResponseErr(ctx, rr, e.New(4001, "bad", "x"))
	assert.Equal(t, DefaultHttpErrorCode, rr.Code)

	var resp Response
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 418, resp.Status)
	assert.Equal(t, 4001, resp.Error.Code)
}

func TestHttpResponse_NoTrace(t *testing.T) {
	rr := httptest.NewRecorder()
	HttpResponse(context.Background(), rr, 200, map[string]string{"hello": "world"})
	assert.Equal(t, 200, rr.Code)

	var resp Response
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Status)
	assert.Empty(t, resp.TraceID)
}

func TestHttpResponse_WithMockTracer(t *testing.T) {
	// mocktracer span context is NOT a jaeger.SpanContext so TraceID will be empty.
	tracer := mocktracer.New()
	span := tracer.StartSpan("test")
	defer span.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	rr := httptest.NewRecorder()
	HttpResponse(ctx, rr, 200, nil)

	var resp Response
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.TraceID)
}

func TestHttpResponse_WithJaegerSpan(t *testing.T) {
	tracer, closer := jaeger.NewTracer(
		"test",
		jaeger.NewConstSampler(true),
		jaeger.NewNullReporter(),
	)
	defer closer.Close()
	span := tracer.StartSpan("op")
	defer span.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	rr := httptest.NewRecorder()
	HttpResponse(ctx, rr, 200, map[string]string{"ok": "yes"})

	var resp Response
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.TraceID)
}
