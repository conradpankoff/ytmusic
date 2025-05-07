package ytdirect

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"

	"fknsrs.biz/p/ytmusic/internal/ctxhttpclient"
	"fknsrs.biz/p/ytmusic/internal/httpcache"
)

func withContext(ctx context.Context, fn func(ctx context.Context)) error {
	cachePath := os.Getenv("TEST_CACHE_PATH")
	if cachePath != "" {
		cacheDB, err := bbolt.Open(cachePath, 0600, nil)
		if err != nil {
			return err
		}
		defer cacheDB.Close()

		ctx = ctxhttpclient.WithHTTPClient(ctx, &http.Client{
			Transport: httpcache.NewTransport(nil, httpcache.NewBBoltStorage(cacheDB), 0),
		})
	}

	fn(ctx)

	return nil
}

func TestGetChannel(t *testing.T) {
	withContext(context.Background(), func(ctx context.Context) {
		for _, tc := range []struct {
			id  string
			err string
			ch  *Channel
		}{
			{"UCpNvmbdtY8WAzhdNUDxbT2g", "", &Channel{
				ID: "UCpNvmbdtY8WAzhdNUDxbT2g",
				Title: "Taylor Lee Czer - Topic",
			}},
		} {
			t.Run(tc.id, func(t *testing.T) {
				a := assert.New(t)

				v, err := GetChannel(ctx, tc.id)
				if tc.err == "" {
					a.NoError(err)
					if a.NotNil(v) {
						a.Equal(tc.ch, v)
					}
				} else {
					a.Nil(v)
					if a.Error(err) {
						a.Contains(err.Error(), tc.err)
					}
				}
			})
		}
	})
}
