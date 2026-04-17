// Author: Daniel TAN
// Date: 2021-09-03 12:24:12
// LastEditors: Daniel TAN
// LastEditTime: 2021-10-22 01:11:57
// FilePath: /trinity-micro/core/requests/requests.go
// Description:
package requests

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	Username string `json:"username" xml:"Username"`
	Age      int    `json:"age" xml:"Age"`
}

func Test_buildBody(t *testing.T) {

	// test json
	{
		d, err := buildBody(test{Username: "test1", Age: 1}, nil)
		assert.Equal(t, nil, err, "has err")
		buf := &bytes.Buffer{}
		buf.ReadFrom(d)
		data := buf.Bytes()
		assert.Equal(t, nil, err, "has err")
		assert.Equal(t, "{\"username\":\"test1\",\"age\":1}", string(data))
	}
	// test xml
	{
		d, err := buildBody(test{Username: "test1", Age: 1}, http.Header{HeaderMime: []string{MimeTextXML}})
		assert.Equal(t, nil, err, "has err")
		buf := &bytes.Buffer{}
		buf.ReadFrom(d)
		data := buf.Bytes()
		assert.Equal(t, nil, err, "has err")
		assert.Equal(t, "<test><Username>test1</Username><Age>1</Age></test>", string(data))
	}
	// test json
	{
		d, err := buildBody(test{Username: "test1", Age: 1}, http.Header{HeaderMime: []string{MimeJson}})
		assert.Equal(t, nil, err, "has err")
		buf := &bytes.Buffer{}
		buf.ReadFrom(d)
		data := buf.Bytes()
		assert.Equal(t, nil, err, "has err")
		assert.Equal(t, "{\"username\":\"test1\",\"age\":1}", string(data))
	}
	// []byte body
	{
		d, err := buildBody([]byte(`raw-bytes`), nil)
		assert.NoError(t, err)
		buf := &bytes.Buffer{}
		buf.ReadFrom(d)
		assert.Equal(t, "raw-bytes", buf.String())
	}
	// io.Reader body
	{
		d, err := buildBody(strings.NewReader("reader-body"), nil)
		assert.NoError(t, err)
		buf := &bytes.Buffer{}
		buf.ReadFrom(d)
		assert.Equal(t, "reader-body", buf.String())
	}
	// nil body
	{
		d, err := buildBody(nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, d)
	}
	// xml marshal error
	{
		_, err := buildBody(make(chan int), http.Header{HeaderMime: []string{MimeXML}})
		assert.Error(t, err)
	}
	// json marshal error
	{
		_, err := buildBody(make(chan int), nil)
		assert.Error(t, err)
	}
}

func TestNewRequest(t *testing.T) {
	r := NewRequest()
	assert.NotNil(t, r)
}

func TestHttpRequest_Call_JSONSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderMime, MimeJson)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"username":"alice","age":30}`))
	}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest)
	assert.NoError(t, err)
	assert.Equal(t, "alice", dest.Username)
	assert.Equal(t, 30, dest.Age)
}

func TestHttpRequest_Call_XMLSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderMime, MimeXML+"; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<test><Username>bob</Username><Age>40</Age></test>`))
	}))
	defer server.Close()

	r := NewRequest()
	header := http.Header{HeaderMime: []string{MimeXML}}
	var dest test
	err := r.Call(context.Background(), http.MethodPost, server.URL, header, test{Username: "bob", Age: 40}, &dest)
	assert.NoError(t, err)
	assert.Equal(t, "bob", dest.Username)
}

func TestHttpRequest_Call_HTMLUnsupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderMime, MimeTextHTML)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html></html>`))
	}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "html")
}

func TestHttpRequest_Call_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderMime, MimeJson)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest)
	assert.Error(t, err)
}

func TestHttpRequest_Call_XMLDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderMime, MimeXML)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not-xml`))
	}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest)
	assert.Error(t, err)
}

func TestHttpRequest_Call_BadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderMime, MimeJson)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"username":"alice","age":30}`))
	}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest)
	assert.Error(t, err)
}

func TestHttpRequest_Call_BadURL(t *testing.T) {
	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, "://invalid-url", nil, nil, &dest)
	assert.Error(t, err)
}

func TestHttpRequest_Call_BadMethod(t *testing.T) {
	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), "bad method", "http://example.com", nil, nil, &dest)
	assert.Error(t, err)
}

func TestHttpRequest_Call_BodyError(t *testing.T) {
	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, "http://example.com", nil, make(chan int), &dest)
	assert.Error(t, err)
}

func TestHttpRequest_Call_InterceptorError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest,
		func(req *http.Request) error {
			return errors.New("interceptor failed")
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interceptor")
}

func TestHttpRequest_Call_InterceptorSuccess(t *testing.T) {
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")
		w.Header().Set(HeaderMime, MimeJson)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"username":"x","age":1}`))
	}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest,
		func(req *http.Request) error {
			req.Header.Set("X-Custom", "value")
			return nil
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "value", gotHeader)
}

func TestHttpRequest_Call_DoError(t *testing.T) {
	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, "http://127.0.0.1:1", nil, nil, &dest)
	assert.Error(t, err)
}

func TestHttpRequest_Call_ReadBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("response writer does not implement hijacker")
		}
		conn, bw, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		bw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\nshort")
		bw.Flush()
		conn.Close()
	}))
	defer server.Close()

	r := NewRequest()
	var dest test
	err := r.Call(context.Background(), http.MethodGet, server.URL, nil, nil, &dest)
	assert.Error(t, err)
}
