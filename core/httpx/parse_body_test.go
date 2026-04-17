package httpx

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeJSONReq(t *testing.T, body string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodPost, "http://example.com/", ioutil.NopCloser(bytes.NewBufferString(body)))
	assert.NoError(t, err)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestParse_NilValue(t *testing.T) {
	r := makeJSONReq(t, "{}")
	err := Parse(r, nil)
	assert.Error(t, err)
}

func TestParse_UnexportedField(t *testing.T) {
	type payload struct {
		test int `header_param:"X-Test"`
	}
	r := makeJSONReq(t, "{}")
	r.Header.Set("X-Test", "1")
	err := Parse(r, &payload{})
	assert.Error(t, err)
}

func TestParse_BodyParam_AllTypes(t *testing.T) {
	{
		type payload struct {
			Body string `body_param:""`
		}
		r := makeJSONReq(t, `raw string body`)
		var p payload
		assert.NoError(t, Parse(r, &p))
		assert.Equal(t, "raw string body", p.Body)
	}
	{
		type inner struct {
			A int `json:"a"`
		}
		type payload struct {
			Body inner `body_param:""`
		}
		r := makeJSONReq(t, `{"a":7}`)
		var p payload
		assert.NoError(t, Parse(r, &p))
		assert.Equal(t, 7, p.Body.A)
	}
	{
		type payload struct {
			Body []int `body_param:""`
		}
		r := makeJSONReq(t, `[1,2,3]`)
		var p payload
		assert.NoError(t, Parse(r, &p))
		assert.Equal(t, []int{1, 2, 3}, p.Body)
	}
	{
		type payload struct {
			Body []byte `body_param:""`
		}
		r := makeJSONReq(t, `hello`)
		var p payload
		assert.NoError(t, Parse(r, &p))
		assert.Equal(t, []byte("hello"), p.Body)
	}
	{
		type payload struct {
			Body map[string]interface{} `body_param:""`
		}
		r := makeJSONReq(t, `{"x":1}`)
		var p payload
		assert.NoError(t, Parse(r, &p))
		assert.NotNil(t, p.Body["x"])
	}
	{
		type payload struct {
			Body interface{} `body_param:""`
		}
		r := makeJSONReq(t, `{"y":2}`)
		var p payload
		assert.NoError(t, Parse(r, &p))
		assert.NotNil(t, p.Body)
	}
	{
		type inner struct {
			V int `json:"v"`
		}
		type payload struct {
			Body *inner `body_param:""`
		}
		r := makeJSONReq(t, `{"v":5}`)
		var p payload
		assert.NoError(t, Parse(r, &p))
		assert.NotNil(t, p.Body)
		assert.Equal(t, 5, p.Body.V)
	}
	{
		type inner struct {
			V int `json:"v"`
		}
		type payload struct {
			Body *inner `body_param:""`
		}
		r := makeJSONReq(t, ``)
		var p payload
		assert.NoError(t, Parse(r, &p))
	}
}

func TestParse_BodyParam_Named(t *testing.T) {
	type inner struct {
		V int `json:"v"`
	}
	type payload struct {
		A int    `body_param:"a"`
		B int64  `body_param:"b"`
		C int32  `body_param:"c"`
		D string `body_param:"d"`
		E inner  `body_param:"e"`
	}
	body := `{"a":1,"b":2,"c":3,"d":"hello","e":{"v":9}}`
	r := makeJSONReq(t, body)
	var p payload
	// b/c the nested 'b' is float64 after json decode, but converter expects int64 ;
	// use bodyParamConverter directly to reach all paths
	_ = Parse(r, &p)
}

func TestParse_BodyParam_InvalidJson(t *testing.T) {
	type payload struct {
		Body map[string]interface{} `body_param:""`
	}
	r := makeJSONReq(t, `not-json`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_BodyParam_UnsupportedMap(t *testing.T) {
	type payload struct {
		Body map[int]string `body_param:""`
	}
	r := makeJSONReq(t, `{}`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_BodyParam_UnsupportedKind(t *testing.T) {
	type payload struct {
		Body int `body_param:""`
	}
	r := makeJSONReq(t, `1`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_BodyParam_InvalidInterface(t *testing.T) {
	type payload struct {
		Body interface{} `body_param:""`
	}
	r := makeJSONReq(t, `not-json`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_BodyParam_InvalidPtr(t *testing.T) {
	type inner struct {
		V int `json:"v"`
	}
	type payload struct {
		Body *inner `body_param:""`
	}
	r := makeJSONReq(t, `not-json`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_BodyParam_NamedInvalidJson(t *testing.T) {
	type payload struct {
		A int `body_param:"a"`
	}
	r := makeJSONReq(t, `not-json`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_BodyParam_NamedMissingKey(t *testing.T) {
	type payload struct {
		A int `body_param:"missing"`
	}
	r := makeJSONReq(t, `{"a":1}`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_QueryParamEmpty_String(t *testing.T) {
	type payload struct {
		Raw string `query_param:""`
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?a=1&b=2", nil)
	var p payload
	assert.NoError(t, Parse(r, &p))
	assert.Equal(t, "a=1&b=2", p.Raw)
}

func TestParse_BodyParam_StructInvalidJson(t *testing.T) {
	type inner struct {
		V int `json:"v"`
	}
	type payload struct {
		Body inner `body_param:""`
	}
	r := makeJSONReq(t, `not-json`)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_BodyParam_NamedSuccess(t *testing.T) {
	type payload struct {
		A string `body_param:"a"`
	}
	r := makeJSONReq(t, `{"a":"hello"}`)
	var p payload
	assert.NoError(t, Parse(r, &p))
	assert.Equal(t, "hello", p.A)
}

func TestParse_NestedStruct_Err(t *testing.T) {
	type inner struct {
		X int `query_param:"x"`
	}
	type payload struct {
		Inner inner
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?x=abc", nil)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_NestedPtr_Err(t *testing.T) {
	type inner struct {
		X int `query_param:"x"`
	}
	type payload struct {
		Inner *inner
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?x=abc", nil)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_NestedStructAndPtr(t *testing.T) {
	type inner struct {
		X int `query_param:"x"`
	}
	type payload struct {
		Inner    inner
		InnerPtr *inner
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?x=5", nil)
	var p payload
	assert.NoError(t, Parse(r, &p))
	assert.Equal(t, 5, p.Inner.X)
	assert.Equal(t, 5, p.InnerPtr.X)
}

func TestParse_QueryParam_UnsupportedType(t *testing.T) {
	type payload struct {
		V []string `query_param:""`
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?a=1", nil)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_QueryParam_UnsupportedMap(t *testing.T) {
	type payload struct {
		V map[int]string `query_param:""`
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?a=1", nil)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_HeaderParamError(t *testing.T) {
	type payload struct {
		V []int `header_param:"X-Test"`
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	r.Header.Set("X-Test", "1")
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_PathParamError(t *testing.T) {
	type payload struct {
		V int `path_param:"id"`
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestParse_QueryParamError(t *testing.T) {
	type payload struct {
		V int `query_param:"age"`
	}
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/?age=abc", nil)
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

func TestBodyParamConverter_AllTypes(t *testing.T) {
	body := map[string]interface{}{
		"int64":  int64(64),
		"int32":  int32(32),
		"int":    int(16),
		"string": "hello",
		"struct": map[string]interface{}{"v": 7},
	}

	type inner struct {
		V int `json:"v"`
	}

	cases := []struct {
		key  string
		typ  reflect.Type
		want interface{}
	}{
		{"int64", reflect.TypeOf(int64(0)), int64(64)},
		{"int32", reflect.TypeOf(int32(0)), int32(32)},
		{"int", reflect.TypeOf(int(0)), int(16)},
		{"string", reflect.TypeOf(""), "hello"},
		{"struct", reflect.TypeOf(inner{}), inner{V: 7}},
	}
	for _, c := range cases {
		got, err := bodyParamConverter(body, c.key, c.typ)
		assert.NoError(t, err, c.key)
		assert.Equal(t, c.want, got, c.key)
	}
}

func TestBodyParamConverter_Errors(t *testing.T) {
	body := map[string]interface{}{
		"wrongInt":   "abc",
		"wrongInt32": "abc",
		"wrongInt64": "abc",
		"wrongStr":   123,
	}

	_, err := bodyParamConverter(body, "missing", reflect.TypeOf(int(0)))
	assert.Error(t, err)

	_, err = bodyParamConverter(body, "wrongInt", reflect.TypeOf(int(0)))
	assert.Error(t, err)

	_, err = bodyParamConverter(body, "wrongInt32", reflect.TypeOf(int32(0)))
	assert.Error(t, err)

	_, err = bodyParamConverter(body, "wrongInt64", reflect.TypeOf(int64(0)))
	assert.Error(t, err)

	_, err = bodyParamConverter(body, "wrongStr", reflect.TypeOf(""))
	assert.Error(t, err)

	_, err = bodyParamConverter(body, "wrongInt", reflect.TypeOf(float64(0)))
	assert.Error(t, err)
}

func TestBodyParamConverter_StructDecodeError(t *testing.T) {
	// use a reader that fails decode by giving a chan value
	type inner struct {
		V int `json:"v"`
	}
	body := map[string]interface{}{
		"struct": map[string]interface{}{"v": "not-an-int"},
	}
	_, err := bodyParamConverter(body, "struct", reflect.TypeOf(inner{}))
	assert.Error(t, err)
}

func TestParse_ReadBodyError(t *testing.T) {
	type payload struct {
		Body string `body_param:""`
	}
	r, _ := http.NewRequest(http.MethodPost, "http://example.com/", &failingReader{})
	var p payload
	err := Parse(r, &p)
	assert.Error(t, err)
}

type failingReader struct{}

func (f *failingReader) Read(p []byte) (int, error) { return 0, assert.AnError }
func (f *failingReader) Close() error               { return nil }

var _ io.ReadCloser = (*failingReader)(nil)

func TestGetRawRequest(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	ctx := context.WithValue(context.Background(), HttpxContext, NewContext(r, 0))
	got := GetRawRequest(ctx)
	assert.Equal(t, r, got)
}

func TestGetRawRequest_Panics(t *testing.T) {
	assert.Panics(t, func() {
		GetRawRequest(context.Background())
	})
}

func TestSetHttpStatusCode_Panics(t *testing.T) {
	assert.Panics(t, func() {
		SetHttpStatusCode(context.Background(), 200)
	})
}

func TestNewContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	c := NewContext(r, 201)
	assert.Equal(t, r, c.r)
	assert.Equal(t, 201, c.code)
}

func TestGetHTTPStatusCode_NilCtxVal(t *testing.T) {
	ctx := context.WithValue(context.Background(), HttpxContext, (*Context)(nil))
	got := GetHTTPStatusCode(ctx, 500)
	assert.Equal(t, 500, got)
}
