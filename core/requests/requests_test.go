// Author: Daniel TAN
// Date: 2021-09-03 12:24:12
// LastEditors: Daniel TAN
// LastEditTime: 2021-10-22 01:11:57
// FilePath: /trinity-micro/core/requests/requests.go
// Description:
package requests

import (
	"bytes"
	"net/http"
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

}
