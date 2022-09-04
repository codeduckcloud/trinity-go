/*
 * @Author: Daniel TAN
 * @Description:
 * @Date: 2021-08-06 09:31:01
 * @LastEditTime: 2021-09-02 17:59:24
 * @LastEditors: Daniel TAN
 * @FilePath: /fr-price-common-pkg/core/httpx/parse_test.go
 */
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/codeduckcloud/trinity-go/core/requests"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var (
	a = func(test string) bool {
		return false
	}
	b = func(test, test2, test3 string) bool {
		return false
	}

	c = func(test, asdasf, wqe string, d int) bool {
		return false
	}
	d = func() bool {
		return false
	}
	h = func(Args struct {
		Test string
	}) bool {
		return false
	}
	f = func(ctx context.Context, Args struct {
		Test string
	}) bool {
		return false
	}
	g = func(ctx context.Context, Args struct {
		Test  string `query_param:"test"`
		Test2 string `query_param:"test2"`
	}) bool {
		return false
	}
	ctxTest = context.WithValue(context.Background(), "test", "123")
)

func TestIsHandler(t *testing.T) {

	type args struct {
		handlerType reflect.Type
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				handlerType: reflect.TypeOf(a),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHandler(tt.args.handlerType); got != tt.want {
				t.Errorf("IsHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandlerNumsIn(t *testing.T) {
	type args struct {
		handlerType reflect.Type
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				handlerType: reflect.TypeOf(a),
			},
			want: 1,
		},
		{
			name: "1",
			args: args{
				handlerType: reflect.TypeOf(b),
			},
			want: 3,
		},
		{
			name: "1",
			args: args{
				handlerType: reflect.TypeOf(c),
			},
			want: 4,
		},
		{
			name: "1",
			args: args{
				handlerType: reflect.TypeOf(d),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandlerNumsIn(tt.args.handlerType); got != tt.want {
				t.Errorf("HandlerNumsIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidHandler(t *testing.T) {
	type args struct {
		r           *http.Request
		handlerType reflect.Type
	}
	tests := []struct {
		name       string
		args       args
		want       []reflect.Value
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				handlerType: reflect.TypeOf("123"),
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "wrong handler type , must be func ",
		},
		{
			name: "1",
			args: args{
				handlerType: reflect.TypeOf(a),
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "wrong handler , unsupported type ",
		},
		{
			name: "2",
			args: args{
				handlerType: reflect.TypeOf(b),
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "wrong handler , unsupported type ",
		},
		{
			name: "3",
			args: args{
				handlerType: reflect.TypeOf(c),
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "wrong handler , unsupported type ",
		},
		{
			name: "3",
			args: args{
				r:           (*http.Request)(&http.Request{}).WithContext(ctxTest),
				handlerType: reflect.TypeOf(func(ctx interface{}) {}),
			},
			want:    []reflect.Value{reflect.ValueOf(ctxTest)},
			wantErr: false,
		},
		{
			name: "3",
			args: args{
				r: (*http.Request)(&http.Request{}).WithContext(ctxTest),
				handlerType: reflect.TypeOf(func(ctx interface {
					test()
				}) {
				}),
			},
			wantErr:    true,
			wantErrMsg: "wrong handler , interface only support context",
		},
		{
			name: "4",
			args: args{
				handlerType: reflect.TypeOf(d),
			},
			want:    make([]reflect.Value, 0),
			wantErr: false,
		},
		{
			name: "5",
			args: args{
				handlerType: reflect.TypeOf(h),
			},
			want:    []reflect.Value{reflect.ValueOf(struct{ Test string }{})},
			wantErr: false,
		},
		{
			name: "6",
			args: args{
				r:           (*http.Request)(&http.Request{}).WithContext(ctxTest),
				handlerType: reflect.TypeOf(f),
			},
			want: []reflect.Value{
				reflect.ValueOf(ctxTest), reflect.ValueOf(struct{ Test string }{})},
			wantErr: false,
		},
		{
			name: "7",
			args: args{
				r: (*http.Request)(&http.Request{
					Method: "POST",
					URL: &url.URL{
						RawQuery: "test=123&test2=354",
					},
				}).WithContext(ctxTest),
				handlerType: reflect.TypeOf(g),
			},
			want: []reflect.Value{
				reflect.ValueOf(ctxTest), reflect.ValueOf(struct {
					Test  string `query_param:"test"`
					Test2 string `query_param:"test2"`
				}{
					Test:  "123",
					Test2: "354",
				})},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InvokeHandler(tt.args.handlerType, tt.args.r)
			if tt.wantErr {
				if err == nil {
					assert.FailNow(t, "expect has error actual not ")
				}
				assert.Equal(t, tt.wantErrMsg, err.Error(), "wrong err msg")
			} else {
				for i := range tt.want {
					assert.Equal(t, tt.want[i].Interface(), got[i].Interface(), "wrong got ")
				}

			}

		})
	}
}

func Test_getHTTPStatusCode(t *testing.T) {
	type args struct {
		ctx           context.Context
		defaultStatus int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				ctx:           context.Background(),
				defaultStatus: 999,
			},
			want: 999,
		},
		{
			name: "1",
			args: args{
				ctx:           context.WithValue(context.Background(), HttpxContext, &Context{code: 200}),
				defaultStatus: 999,
			},
			want: 200,
		},
		{
			name: "1",
			args: args{
				ctx:           context.WithValue(context.Background(), HttpxContext, &Context{code: 123}),
				defaultStatus: 999,
			},
			want: 123,
		},
		{
			name: "1",
			args: args{
				ctx:           context.WithValue(context.Background(), HttpxContext, &Context{}),
				defaultStatus: 999,
			},
			want: 999,
		},
		{
			name: "1",
			args: args{
				ctx:           context.WithValue(context.Background(), HttpxContext, &Context{}),
				defaultStatus: 999,
			},
			want: 999,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHTTPStatusCode(tt.args.ctx, tt.args.defaultStatus); got != tt.want {
				t.Errorf("getHTTPStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetHttpStatusCode(t *testing.T) {
	type args struct {
		ctx    context.Context
		status int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				ctx:    context.WithValue(context.Background(), HttpxContext, &Context{code: 999}),
				status: 200,
			},
			want: 200,
		},
		{
			name: "1",
			args: args{
				ctx:    context.WithValue(context.Background(), HttpxContext, &Context{code: 999}),
				status: 777,
			},
			want: 777,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetHttpStatusCode(tt.args.ctx, tt.args.status)
			if got := GetHTTPStatusCode(tt.args.ctx, 999); got != tt.want {
				t.Errorf("getHTTPStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getPtrInt(a int) *int {
	return &a

}

func TestParse(t *testing.T) {
	type test struct {
		ID            string                 `path_param:"id"`
		XXX           string                 `header_param:"test"`
		Age           int                    `query_param:"age"`
		Percentage    string                 `query_param:"percent"`
		Percentage2   float64                `query_param:"percent"`
		Query         url.Values             `query_param:""`
		Query1        map[string]string      `query_param:""`
		Query2        map[string]interface{} `query_param:""`
		Query3        map[string][]string    `query_param:""`
		NotExistQuery int                    `query_param:"not_existttt"`
		NotExistParam int                    `header_param:"not_existttt"`
	}
	type args struct {
		r *http.Request
		v interface{}
	}
	tests := []struct {
		name       string
		args       args
		want       interface{}
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "1",
			args: args{
				r: newReq(&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"id"},
						Values: []string{"a"},
					},
				}, "GET", "http://hello.com/a?name=hello&age=18&percent=3.4", nil, nil),
			},
			wantErr:    true,
			wantErrMsg: "parsing error , empty value to parse",
		},
		{
			name: "2",
			args: args{
				r: newReq(&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"id"},
						Values: []string{"a"},
					},
				}, "GET", "http://hello.com/a?name=hello&age=18&percent=3.4", nil, map[string]string{
					"test": "234",
				}),
				v: &test{},
			},
			want: &test{
				ID:          "a",
				XXX:         "234",
				Age:         18,
				Percentage:  "3.4",
				Percentage2: 3.4,
				Query: url.Values{
					"age":     []string{"18"},
					"name":    []string{"hello"},
					"percent": []string{"3.4"},
				},
				Query1: map[string]string{
					"age":     "18",
					"name":    "hello",
					"percent": "3.4",
				},
				Query2: map[string]interface{}{
					"age":     "18",
					"name":    "hello",
					"percent": "3.4",
				},
				Query3: map[string][]string{
					"age":     {"18"},
					"name":    {"hello"},
					"percent": {"3.4"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Parse(tt.args.r, tt.args.v)
			if err != nil || tt.wantErr {
				assert.Equal(t, tt.wantErrMsg, err.Error(), "wrong error ")
			} else {
				assert.Equal(t, nil, err, "wrong ")
				assert.Equal(t, tt.want, tt.args.v, "wrong ")
			}
		})
	}
}

func TestParseValidator(t *testing.T) {
	type test struct {
		ID         int64   `path_param:"id" validate:"eq=3"`
		XXX        int64   `header_param:"test"`
		Age        int     `query_param:"age"`
		Percentage float64 `query_param:"percent"`
	}
	type args struct {
		r *http.Request
		v interface{}
	}
	tests := []struct {
		name       string
		args       args
		want       interface{}
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "1",
			args: args{
				r: newReq(&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"id"},
						Values: []string{"123"},
					},
				}, "GET", "http://hello.com/123?name=hello&age=18&percent=3.4", nil, map[string]string{
					"test": "234",
				}),
				v: &test{},
			},
			wantErr:    true,
			wantErrMsg: "httpx.Parse validate error, err: Key: 'test.ID' Error:Field validation for 'ID' failed on the 'eq' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Parse(tt.args.r, tt.args.v)
			if err != nil || tt.wantErr {
				assert.Equal(t, tt.wantErrMsg, err.Error(), "wrong error ")
			} else {
				assert.Equal(t, nil, err, "wrong ")
				assert.Equal(t, tt.want, tt.args.v, "wrong ")
			}
		})
	}
}

func newReq(chiCtx *chi.Context, method string, url string, body interface{}, header map[string]string) *http.Request {
	var bodyTemp io.Reader
	if body != nil {
		r, ok := body.(io.Reader)
		if ok {
			bodyTemp = r
		} else {
			var bodyBytes []byte
			mime := header[requests.HeaderMime]
			switch mime {
			case requests.MimeXML, requests.MimeTextXML:
				bodyBytes, _ = xml.Marshal(body)
			default:
				bodyBytes, _ = json.Marshal(body)
			}
			bodyTemp = bytes.NewReader(bodyBytes)
		}
	}
	r, _ := http.NewRequest(method, url, bodyTemp)
	for k, v := range header {
		r.Header.Add(k, v)
	}
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	return r
}
