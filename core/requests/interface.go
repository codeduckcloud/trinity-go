// Author: Daniel TAN
// Date: 2021-10-02 01:05:20
// LastEditors: Daniel TAN
// LastEditTime: 2021-10-02 01:06:29
// FilePath: /trinity-micro/core/requests/interface.go
// Description:
package requests

import (
	"context"
	"net/http"
)

type Requests interface {
	Call(ctx context.Context, method string, url string, header http.Header, body interface{}, dest interface{}, requestInterceptors ...Interceptor) error
}
