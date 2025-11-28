package httpcache

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"fknsrs.biz/p/ytmusic/internal/ctxclock"
	"fknsrs.biz/p/ytmusic/internal/ctxlogger"
)

type cachedResponse struct {
	UpdatedAt  time.Time
	URL        string
	Status     string
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (r *cachedResponse) makeResponse(req *http.Request) *http.Response {
	return &http.Response{
		Status:        r.Status,
		StatusCode:    r.StatusCode,
		Header:        r.Header,
		Body:          io.NopCloser(bytes.NewReader(r.Body)),
		ContentLength: int64(len(r.Body)),
		Request:       req,
	}
}

type Storage interface {
	Fetch(ctx context.Context, u *url.URL) (*cachedResponse, error)
	Save(ctx context.Context, u *url.URL, res *http.Response) (*cachedResponse, error)
}

var bboltBucketName = []byte("cache")

type BBoltStorage struct {
	db *bbolt.DB
}

func NewBBoltStorage(db *bbolt.DB) *BBoltStorage {
	return &BBoltStorage{db: db}
}

func makeBBoltKey(u *url.URL) []byte {
	h := sha1.New()
	io.WriteString(h, u.String())
	return []byte(filepath.Join(u.Host, hex.EncodeToString(h.Sum(nil))))
}

func (s *BBoltStorage) Fetch(ctx context.Context, u *url.URL) (*cachedResponse, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(bboltBucketName)
	if b == nil {
		return nil, nil
	}

	d := b.Get(makeBBoltKey(u))
	if d == nil {
		return nil, nil
	}

	if err := tx.Rollback(); err != nil {
		return nil, err
	}

	var r cachedResponse
	if err := gob.NewDecoder(bytes.NewReader(d)).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *BBoltStorage) Save(ctx context.Context, u *url.URL, res *http.Response) (*cachedResponse, error) {
	d, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r := cachedResponse{
		UpdatedAt:  ctxclock.MustNow(ctx),
		URL:        u.String(),
		Status:     res.Status,
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       d,
	}

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(r); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists(bboltBucketName)
	if err != nil {
		return nil, err
	}

	if err := b.Put(makeBBoltKey(u), buf.Bytes()); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &r, nil
}

type Transport struct {
	transport http.RoundTripper
	storage   Storage
	maxAge    time.Duration
}

func NewTransport(transport http.RoundTripper, storage Storage, maxAge time.Duration) *Transport {
	if transport == nil {
		transport = http.DefaultTransport
	}

	if maxAge == 0 {
		maxAge = time.Hour * 24
	}

	return &Transport{
		transport: transport,
		storage:   storage,
		maxAge:    maxAge,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return t.transport.RoundTrip(req)
	}

	if cr, err := t.storage.Fetch(req.Context(), req.URL); err != nil {
		return nil, fmt.Errorf("httpcache.Transport.RoundTrip: %w", err)
	} else if cr == nil {
		ctxlogger.GetLogger(req.Context()).Debug("httpcache: miss")
	} else {
		now := ctxclock.MustNow(req.Context())
		earliest := now.Add(0 - t.maxAge)
		age := now.Sub(cr.UpdatedAt)

		l := ctxlogger.GetLogger(req.Context()).WithFields(logrus.Fields{
			"httpcache.response.time_now":      now,
			"httpcache.response.time_earliest": earliest,
			"httpcache.response.time_updated":  cr.UpdatedAt,
			"httpcache.response.age_max":       t.maxAge,
			"httpcache.response.age_actual":    age,
		})

		if cr.UpdatedAt.Before(earliest) {
			l.Debug("httpcache: stale")
		} else {
			l.Debug("httpcache: hit")
			return cr.makeResponse(req), nil
		}
	}

	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return res, nil
	}

	cr, err := t.storage.Save(req.Context(), req.URL, res)
	if err != nil {
		return nil, err
	}

	return cr.makeResponse(req), nil
}
