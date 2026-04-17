package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type diTestService struct {
	name string
}

func (s *diTestService) NoArgs() {}

func (s *diTestService) WithCtx(ctx context.Context) {}

func (s *diTestService) WithWriter(w http.ResponseWriter) {
	if w != nil {
		w.WriteHeader(202)
	}
}

func (s *diTestService) WithRequest(r *http.Request) {}

func (s *diTestService) WithSelf(self *diTestService) {}

func (s *diTestService) WithStruct(args struct {
	Age int `query_param:"age"`
}) {
}

func (s *diTestService) WithStructPtr(args *struct {
	Age int `query_param:"age"`
}) {
}

type unsupportedIface interface{ Foo() }

func (s *diTestService) WithUnsupportedIface(i unsupportedIface) {}

func (s *diTestService) WithBadStruct(args struct {
	Age int `query_param:"age"`
}) {
}

func TestDIParamHandler_NoArgsReturnsVoid(t *testing.T) {
	fn := func() {}
	h := DIParamHandler(fn)

	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, 200, rr.Code)
}

func TestDIParamHandler_InvalidHandler_InvokeErr(t *testing.T) {
	fn := func(i int) {}
	h := DIParamHandler(fn)
	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, DefaultHttpErrorCode, rr.Code)
}

func TestDIParamHandler_OneReturn_Err(t *testing.T) {
	fn := func() error { return errors.New("boom") }
	h := DIParamHandler(fn)
	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, DefaultHttpErrorCode, rr.Code)

	var resp Response
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "boom", resp.Error.Message)
}

func TestDIParamHandler_OneReturn_NonErr(t *testing.T) {
	fn := func() string { return "ok" }
	h := DIParamHandler(fn)
	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, 200, rr.Code)

	var resp Response
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Result)
}

func TestDIParamHandler_OneReturn_NilError(t *testing.T) {
	fn := func() error { return nil }
	h := DIParamHandler(fn)
	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, 200, rr.Code)
}

func TestDIParamHandler_TwoReturn_Err(t *testing.T) {
	fn := func() (string, error) { return "", errors.New("err") }
	h := DIParamHandler(fn)
	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, DefaultHttpErrorCode, rr.Code)
}

func TestDIParamHandler_TwoReturn_OK(t *testing.T) {
	fn := func() (string, error) { return "ok", nil }
	h := DIParamHandler(fn)
	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, 200, rr.Code)
}

func TestDIParamHandler_TooManyReturns(t *testing.T) {
	fn := func() (string, string, error) { return "a", "b", nil }
	h := DIParamHandler(fn)
	rr := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	h(rr, r)
	assert.Equal(t, DefaultHttpErrorCode, rr.Code)
}

func TestInvokeMethod_NotFunc(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	_, err := InvokeMethod(reflect.TypeOf("hello"), r, nil, httptest.NewRecorder())
	assert.Error(t, err)
}

func TestInvokeMethod_Context(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	fn := func(ctx context.Context) {}
	vals, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.NoError(t, err)
	assert.Len(t, vals, 1)
}

func TestInvokeMethod_Writer(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	rr := httptest.NewRecorder()
	fn := func(w http.ResponseWriter) {}
	vals, err := InvokeMethod(reflect.TypeOf(fn), r, nil, rr)
	assert.NoError(t, err)
	assert.Len(t, vals, 1)
}

func TestInvokeMethod_UnsupportedInterface(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	fn := func(i unsupportedIface) {}
	_, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.Error(t, err)
}

func TestInvokeMethod_Struct(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?age=10", nil)
	fn := func(args struct {
		Age int `query_param:"age"`
	}) {
	}
	vals, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.NoError(t, err)
	assert.Len(t, vals, 1)
}

func TestInvokeMethod_Struct_ParseError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?age=abc", nil)
	fn := func(args struct {
		Age int `query_param:"age"`
	}) {
	}
	_, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.Error(t, err)
}

func TestInvokeMethod_Request(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	fn := func(req *http.Request) {}
	vals, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.NoError(t, err)
	assert.Len(t, vals, 1)
}

func TestInvokeMethod_Instance(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	svc := &diTestService{name: "svc"}
	fn := func(s *diTestService) {}
	vals, err := InvokeMethod(reflect.TypeOf(fn), r, svc, httptest.NewRecorder())
	assert.NoError(t, err)
	assert.Len(t, vals, 1)
}

func TestInvokeMethod_PtrStruct(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?age=10", nil)
	fn := func(args *struct {
		Age int `query_param:"age"`
	}) {
	}
	vals, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.NoError(t, err)
	assert.Len(t, vals, 1)
}

func TestInvokeMethod_PtrStruct_ParseError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?age=abc", nil)
	fn := func(args *struct {
		Age int `query_param:"age"`
	}) {
	}
	_, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.Error(t, err)
}

func TestInvokeMethod_UnsupportedKind(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	fn := func(i int) {}
	_, err := InvokeMethod(reflect.TypeOf(fn), r, nil, httptest.NewRecorder())
	assert.Error(t, err)
}

func TestInvokeHandler_NotFunc(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	_, err := InvokeHandler(reflect.TypeOf(1), r)
	assert.Error(t, err)
}

func TestInvokeHandler_StructParseError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?age=abc", nil)
	fn := func(args struct {
		Age int `query_param:"age"`
	}) {
	}
	_, err := InvokeHandler(reflect.TypeOf(fn), r)
	assert.Error(t, err)
}
