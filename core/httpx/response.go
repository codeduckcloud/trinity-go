// Author: Daniel TAN
// Date: 2021-09-05 10:24:33
// LastEditors: Daniel TAN
// LastEditTime: 2021-10-03 14:56:13
// FilePath: /trinity-micro/core/httpx/response.go
// Description:
package httpx

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/codeduckcloud/trinity-go/core/e"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
)

const (
	DefaultHttpErrorCode   int = 400
	DefaultHttpSuccessCode int = 200
)

func JsonResponse(w http.ResponseWriter, status int, res interface{}) {
	j, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(j)
}

func XMLResponse(w http.ResponseWriter, status int, res interface{}) {
	j, err := xml.Marshal(res)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	w.Write(j)
}

type Response struct {
	Status  int         `json:"status" example:"200"`
	Result  interface{} `json:"result,omitempty" `
	Error   *ErrorInfo  `json:"error,omitempty"`
	TraceID string      `json:"trace_id,omitempty" example:"1-trace-it"`
}

type ErrorInfo struct {
	Code    int      `json:"code" example:"400001"`
	Message string   `json:"message" example:"ErrInvalidRequest"`
	Details []string `json:"details" example:"error detail1,error detail2"`
}

func HttpResponseErr(ctx context.Context, w http.ResponseWriter, err error) {
	wrapErr := e.FromErr(err)
	res := &Response{
		Status: GetHTTPStatusCode(ctx, DefaultHttpErrorCode),
		Error: &ErrorInfo{
			Code:    wrapErr.Code(),
			Message: wrapErr.Message(),
			Details: wrapErr.Details(),
		},
	}
	JsonResponse(w, DefaultHttpErrorCode, res)
}

func HttpResponse(ctx context.Context, w http.ResponseWriter, status int, res interface{}) {
	result := &Response{
		Status: status,
		Result: res,
	}
	x := opentracing.SpanFromContext(ctx)
	if x != nil {
		sc, ok := x.Context().(jaeger.SpanContext)
		if ok {
			result.TraceID = sc.TraceID().String()
		}
	}
	JsonResponse(w, status, result)
}
