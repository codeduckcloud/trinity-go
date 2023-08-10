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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Interceptor func(r *http.Request) error

var _ Requests = new(HttpRequest)

type HttpRequest struct {
}

func NewRequest() Requests {
	return &HttpRequest{}
}

// Call
// will return the err when the response code is not >=200 or <= 300
// will decode the response to dest even it return error
func (r *HttpRequest) Call(ctx context.Context, method string, url string, header http.Header, body interface{}, dest interface{}, requestInterceptors ...Interceptor) error {
	if header == nil {
		header = make(http.Header)
	}
	bodyReader, err := buildBody(body, header)
	if err != nil {
		return fmt.Errorf("build body reader error, err: %v", err)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("new request error, err: %v", err)
	}
	req.Close = true
	req.Header = header
	for _, interceptor := range requestInterceptors {
		if err := interceptor(req); err != nil {
			return fmt.Errorf("new request interceptor error, err: %v", err)
		}
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request error, err: %v", err)
	}
	defer resp.Body.Close()
	bodyRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read request body error, err: %v", err)
	}
	resbody := ioutil.NopCloser(bytes.NewReader(bodyRes))
	mime := resp.Header.Get(HeaderMime)
	contextType := strings.Split(mime, ";")[0]
	switch contextType {
	case MimeTextXML, MimeXML:
		if err := xml.NewDecoder(resbody).Decode(dest); err != nil {
			return fmt.Errorf("decode xml error, err: %v, source body : %v", err, string(bodyRes))
		}
	case MimeTextHTML:
		bodyHTML, err := ioutil.ReadAll(resbody)
		if err != nil {
			return fmt.Errorf("read html error, err: %v", err)
		}
		return fmt.Errorf("html unsupported decode to destination, content: %v", string(bodyHTML))
	default:
		if err := json.NewDecoder(resbody).Decode(dest); err != nil {
			return fmt.Errorf("decode json error, err: %v, source body : %v", err, string(bodyRes))
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		 data := ioutil.ReadAll(resbody)
		return fmt.Errorf("actual http response code, actual: %v , error %v ", resp.StatusCode, string(data))
	}

	return nil
}

func buildBody(body interface{}, header http.Header) (io.Reader, error) {
	var bodyTemp io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			bodyTemp = v
		case []byte:
			bodyBytes := v
			bodyTemp = bytes.NewReader(bodyBytes)
		default:
			mime := header.Get(HeaderMime)
			switch mime {
			case MimeXML, MimeTextXML:
				bodyBytes, err := xml.Marshal(body)
				if err != nil {
					return nil, fmt.Errorf("encode xml error, err: %v", err)
				}
				bodyTemp = bytes.NewReader(bodyBytes)
			default:
				bodyBytes, err := json.Marshal(body)
				if err != nil {
					return nil, fmt.Errorf("encode json error, err: %v", err)
				}
				bodyTemp = bytes.NewReader(bodyBytes)
			}
		}
	}
	return bodyTemp, nil
}
