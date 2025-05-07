package ctxtemplate

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"fknsrs.biz/p/ytmusic/internal/templatecollection"
)

// context registration

var collectionKey int

func WithCollection(ctx context.Context, collection templatecollection.Collection) context.Context {
	return context.WithValue(ctx, &collectionKey, collection)
}

func getCollection(ctx context.Context) templatecollection.Collection {
	if v := ctx.Value(&collectionKey); v != nil {
		return v.(templatecollection.Collection)
	}

	return nil
}

var dataKey int

func WithData(ctx context.Context, data map[string]interface{}) context.Context {
	return context.WithValue(ctx, &dataKey, mergeMaps(getData(ctx), data))
}

func getData(ctx context.Context) map[string]interface{} {
	if v := ctx.Value(&dataKey); v != nil {
		return v.(map[string]interface{})
	}

	return nil
}

func mergeMaps(dst, src map[string]interface{}) map[string]interface{} {
	if dst == nil {
		dst = make(map[string]interface{})
	}

	if src != nil {
		for k, v := range src {
			if dstVal, ok := dst[k]; ok {
				if dstMap, ok := dstVal.(map[string]interface{}); ok {
					if srcMap, ok := v.(map[string]interface{}); ok {
						mergeMaps(dstMap, srcMap)
					} else {
						dst[k] = v
					}
				} else {
					dst[k] = v
				}
			} else {
				dst[k] = v
			}
		}
	}

	return dst
}

// middleware

func Register(collection templatecollection.Collection) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithCollection(r.Context(), collection)))
	}
}

// main interface

var (
	ErrNoCollectionInContext = fmt.Errorf("collection not found in context")
)

func ExecuteTemplate(ctx context.Context, wr io.Writer, name string, data map[string]interface{}) error {
	collection := getCollection(ctx)
	if collection == nil {
		return ErrNoCollectionInContext
	}

	if err := collection.ExecuteTemplate(wr, name, mergeMaps(getData(ctx), data)); err != nil {
		return fmt.Errorf("ctxtemplate.ExecuteTemplate: %w", err)
	}

	return nil
}

func ExecuteTemplateIntoResponse(r *http.Request, rw http.ResponseWriter, name string, data map[string]interface{}) error {
	rw.Header().Set("content-type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	return ExecuteTemplate(r.Context(), rw, name, data)
}
