package trinity

import "net/http"

type mux interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Use(middlewares ...func(http.Handler) http.Handler)
	MethodFunc(method, pattern string, handlerFn http.HandlerFunc)
	Head(pattern string, handlerFn http.HandlerFunc)
	Connect(pattern string, handlerFn http.HandlerFunc)
	Options(pattern string, handlerFn http.HandlerFunc)
	Get(pattern string, handlerFn http.HandlerFunc)
	Post(pattern string, handlerFn http.HandlerFunc)
	Put(pattern string, handlerFn http.HandlerFunc)
	Patch(pattern string, handlerFn http.HandlerFunc)
	Delete(pattern string, handlerFn http.HandlerFunc)
	Trace(pattern string, handlerFn http.HandlerFunc)
}
