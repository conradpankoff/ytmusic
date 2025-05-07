package httpcache

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
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
	Fetch(u *url.URL) (*cachedResponse, error)
	Save(u *url.URL, res *http.Response) (*cachedResponse, error)
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

func (s *BBoltStorage) Fetch(u *url.URL) (*cachedResponse, error) {
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

func (s *BBoltStorage) Save(u *url.URL, res *http.Response) (*cachedResponse, error) {
	d, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r := cachedResponse{
		UpdatedAt:  time.Now(),
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

	if cr, err := t.storage.Fetch(req.URL); err == nil && cr != nil && time.Now().Sub(cr.UpdatedAt) < t.maxAge {
		return cr.makeResponse(req), nil
	}

	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return res, nil
	}

	cr, err := t.storage.Save(req.URL, res)
	if err != nil {
		return nil, err
	}

	return cr.makeResponse(req), nil
}
