package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/codeduckcloud/trinity-go/core/httpx"
)

func Recovery() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logMessage := fmt.Sprintf("Recovered from HTTP Request %v %v \n", r.Method, r.URL)
					logMessage += fmt.Sprintf("Trace: %s\n", err)
					logMessage += fmt.Sprintf("\n%s", debug.Stack())
					httpx.HttpResponseErr(r.Context(), w, fmt.Errorf("panic %v", err))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
