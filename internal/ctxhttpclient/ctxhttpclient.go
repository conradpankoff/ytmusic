package ctxhttpclient

import (
	"context"
	"net/http"
)

// context registration

var httpClientKey int

func WithHTTPClient(ctx context.Context, httpClient *http.Client) context.Context {
	return context.WithValue(ctx, &httpClientKey, httpClient)
}

func GetHTTPClient(ctx context.Context) *http.Client {
	if v := ctx.Value(&httpClientKey); v != nil {
		return v.(*http.Client)
	}

	return http.DefaultClient
}

// middleware

func Register(httpClient *http.Client) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithHTTPClient(r.Context(), httpClient)))
	}
}
