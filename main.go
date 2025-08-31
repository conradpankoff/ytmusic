package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"fknsrs.biz/p/sorm"
	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/urfave/negroni/v2"
	"go.etcd.io/bbolt"

	"fknsrs.biz/p/ytmusic/handlers"
	"fknsrs.biz/p/ytmusic/internal/config"
	"fknsrs.biz/p/ytmusic/internal/configreader"
	"fknsrs.biz/p/ytmusic/internal/ctxclock"
	"fknsrs.biz/p/ytmusic/internal/ctxconfig"
	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/ctxhttpclient"
	"fknsrs.biz/p/ytmusic/internal/ctxjobqueue"
	"fknsrs.biz/p/ytmusic/internal/ctxlogger"
	"fknsrs.biz/p/ytmusic/internal/ctxtemplate"
	"fknsrs.biz/p/ytmusic/internal/ctxtimer"
	"fknsrs.biz/p/ytmusic/internal/ffmpeg"
	"fknsrs.biz/p/ytmusic/internal/httpcache"
	"fknsrs.biz/p/ytmusic/internal/jobqueue"
	"fknsrs.biz/p/ytmusic/internal/logrusstackhook"
	"fknsrs.biz/p/ytmusic/internal/ptr"
	"fknsrs.biz/p/ytmusic/internal/queuenames"
	"fknsrs.biz/p/ytmusic/internal/sqlitelogger"
	"fknsrs.biz/p/ytmusic/internal/stringutil"
	"fknsrs.biz/p/ytmusic/internal/templatecollection"
	"fknsrs.biz/p/ytmusic/internal/ytdirect"
	"fknsrs.biz/p/ytmusic/internal/ytdl"
	"fknsrs.biz/p/ytmusic/models"
)

func init() {
	sorm.SetParameterPrefix("?")
}

var cfg = config.Config{
	LogLevel:             logrus.InfoLevel,
	LogDebugLevels:       config.LevelList{logrus.DebugLevel, logrus.TraceLevel},
	LogQueries:           config.LogQueries{Enabled: true, SlowerThan: time.Millisecond * 100},
	LogSORM:              false,
	ApplicationAddr:      ":8080",
	ApplicationDatabase:  "database.db",
	ApplicationCachePath: "cache.db",
	ApplicationDataPath:  "data",
	ApplicationMinify:    true,
	BackgroundWorkers:    1,
}

//go:embed templates
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

func init() {
	for _, configPath := range []string{"config.toml", "config.yaml", "config.yml"} {
		if st, err := os.Stat(configPath); err == nil && st != nil && !st.IsDir() {
			cfg.Config = configPath
		}
	}
}

type simpleQueryLogger struct {
	logger *logrus.Logger
}

func (s *simpleQueryLogger) LogQuery(query string, args []interface{}) {
	fields := logrus.Fields{
		"db.query":      query,
		"db.args.count": len(args),
	}

	for i, e := range args {
		fields[fmt.Sprintf("db.args.%d", i)] = e
	}

	s.logger.WithFields(fields).Info("sorm query start")
}

func (s *simpleQueryLogger) LogQueryAfter(query string, args []interface{}, duration time.Duration, err error) {
	fields := logrus.Fields{
		"db.query":      query,
		"db.duration":   duration,
		"db.error":      err,
		"db.args.count": len(args),
	}

	for i, e := range args {
		fields[fmt.Sprintf("db.args.%d", i)] = e
	}

	s.logger.WithFields(fields).Info("sorm query finish")
}

func main() {
	ctx := context.Background()

	if err := configreader.Read(os.Args[0], os.Args[1:], os.Environ(), &cfg); err != nil {
		panic(err)
	}

	ctx = ctxconfig.WithConfig(ctx, cfg)

	fmt.Printf("cfg: %#v\n", cfg.ApplicationDataPath)
	fmt.Printf("cfg context: %#v\n", ctxconfig.GetConfig(ctx).ApplicationDataPath)

	ctx = ctxclock.WithClock(ctx, ctxclock.NewRealClock())

	logger := logrus.New()

	logger.SetLevel(cfg.LogLevel)
	if len(cfg.LogDebugLevels) > 0 {
		logger.AddHook(logrusstackhook.NewStackHook(nil, cfg.LogDebugLevels, nil))
	}

	logger.WithFields(logrus.Fields{
		"config.config":                 cfg.Config,
		"config.log_level":              cfg.LogLevel,
		"config.log_debug_levels":       cfg.LogDebugLevels,
		"config.log_queries":            cfg.LogQueries,
		"config.log_sorm":               cfg.LogSORM,
		"config.application_addr":       cfg.ApplicationAddr,
		"config.application_cache_path": cfg.ApplicationCachePath,
		"config.application_database":   cfg.ApplicationDatabase,
		"config.application_data_path":  cfg.ApplicationDataPath,
		"config.application_minify":     cfg.ApplicationMinify,
		"config.background_workers":     cfg.BackgroundWorkers,
	}).Info("program starting")

	if cfg.LogSORM {
		sorm.SetQueryLogger(&simpleQueryLogger{logger})
	}

	ctx = ctxlogger.WithLogger(ctx, logger)

	dbDriver := "sqlite3"

	if !cfg.LogQueries.IsZero() {
		dbDriver = "sqlite3:logged"

		sql.Register(dbDriver, sqlitelogger.New(
			dbDriver,
			&sqlite3.SQLiteDriver{},
			&sqlitelogger.BasicFilter{
				LogSlowerThan: cfg.LogQueries.SlowerThan,
				IgnorePackageStackFrames: []string{
					// standard library
					"database/sql",
					"net/http",
					"runtime",
					// libraries
					"github.com/gorilla/mux",
					"github.com/shogo82148/go-sql-proxy",
					"github.com/urfave/negroni/v2",
					// middleware
					"fknsrs.biz/p/ytmusic/internal/ctxclock",
					"fknsrs.biz/p/ytmusic/internal/ctxdb",
					"fknsrs.biz/p/ytmusic/internal/ctxjobqueue",
					"fknsrs.biz/p/ytmusic/internal/ctxlogger",
					"fknsrs.biz/p/ytmusic/internal/ctxtemplate",
					"fknsrs.biz/p/ytmusic/internal/ctxtimer",
					"fknsrs.biz/p/ytmusic/internal/sqlitelogger",
					// main
					"main",
				},
				IgnoreFunctionQueries: []string{
					"fknsrs.biz/p/ytmusic/internal/jobqueue.(*Worker).Run",
				},
			},
		))
	}

	db, err := sql.Open(dbDriver, cfg.ApplicationDatabase)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx = ctxdb.WithDB(ctx, db)

	cacheDB, err := bbolt.Open(cfg.ApplicationCachePath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer cacheDB.Close()

	ctx = ctxhttpclient.WithHTTPClient(ctx, &http.Client{
		Transport: httpcache.NewTransport(nil, httpcache.NewBBoltStorage(cacheDB), 0),
	})

	ctx = ctxjobqueue.WithWorker(ctx, jobqueue.NewWorker(nil))

	if err := registerJobQueueWorkerFunctions(ctx); err != nil {
		panic(err)
	}

	workers := []worker{
		{
			name: "application",
			run: func(ctx context.Context) error {
				return runApplicationWorker(ctx, cfg.ApplicationAddr)
			},
		},
	}

	for i := 0; i < cfg.BackgroundWorkers; i++ {
		workers = append(workers, worker{
			name: fmt.Sprintf("job_queue.%d", i),
			run: func(ctx context.Context) error {
				return runJobQueueWorker(ctx)
			},
		})
	}

	if err := runAllWorkers(ctx, workers); err != nil {
		panic(err)
	}
}

type worker struct {
	name string
	run  func(ctx context.Context) error
}

func runAllWorkers(ctx context.Context, workers []worker) error {
	done := make(chan error, len(workers))
	cancellers := make([]context.CancelCauseFunc, len(workers))

	var rw sync.RWMutex

	for id, w := range workers {
		go func(id int, w worker) {
			for {
				l := ctxlogger.GetLogger(ctx).WithFields(logrus.Fields{
					"worker.id":   id + 1,
					"worker.name": w.name,
				})

				ctx, cancel := context.WithCancelCause(ctxlogger.WithLogger(ctx, l))

				rw.Lock()
				cancellers[id] = cancel
				rw.Unlock()

				if err := w.run(ctx); err != nil {
					l.WithError(err).Error("worker failed")

					rw.RLock()
					for _, fn := range cancellers {
						if fn == nil {
							continue
						}

						fn(fmt.Errorf("worker %d (%s) failed: %w", id+1, w.name, err))
					}
					rw.RUnlock()
				} else {
					l.Info("worker restarted")
				}

				time.Sleep(time.Second)
			}
		}(id, w)
	}

	var errs []error
	for err := range done {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func directoryExists(name string) bool {
	st, err := os.Stat(name)
	if err != nil {
		return false
	}
	return st.IsDir()
}

type FieldNameValuePair struct {
	Name  string
	Value interface{}
}

func runApplicationWorker(ctx context.Context, addr string) error {
	fmt.Printf("application context config: %#v\n", ctxconfig.GetConfig(ctx).ApplicationDataPath)

	l := ctxlogger.GetLogger(ctx)

	l.WithFields(logrus.Fields{
		"args.addr": addr,
	}).Info("running application worker")

	templateFuncs := template.FuncMap{
		"slice_length": func(v interface{}) int {
			val := reflect.ValueOf(v)
			if val.Kind() != reflect.Slice {
				panic(fmt.Errorf("expected input to be a slice"))
			}
			return val.Len()
		},
		"field_names": func(v interface{}) []string {
			typ := reflect.TypeOf(v)
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if typ.Kind() == reflect.Slice {
				typ = typ.Elem()
			}

			var a []string
			for i := 0; i < typ.NumField(); i++ {
				a = append(a, typ.Field(i).Name)
			}

			return a
		},
		"field_name_value_pairs": func(v interface{}) []FieldNameValuePair {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Ptr {
				val = reflect.Indirect(val)
			}
			if val.Kind() != reflect.Struct {
				panic(fmt.Errorf("expected input value to be a struct"))
			}

			typ := val.Type()

			var a []FieldNameValuePair
			for i := 0; i < typ.NumField(); i++ {
				a = append(a, FieldNameValuePair{typ.Field(i).Name, val.Field(i).Interface()})
			}

			return a
		},
		"first_of": func(a ...interface{}) string {
			for _, e := range a {
				if s := fmt.Sprintf("%v", e); s != "" {
					return s
				}
			}

			return ""
		},
		"format_appropriately": func(obj interface{}, v interface{}) interface{} {
			if v, ok := v.(interface{ FormatAsHTML() template.HTML }); ok {
				return v.FormatAsHTML()
			}

			switch v := v.(type) {
			case time.Time:
				return v.Format(time.RFC3339)
			case *time.Time:
				if v == nil {
					return "never"
				}
				return v.Format(time.RFC3339)
			default:
				return v
			}
		},
		"format_time": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"format_time_null": func(t *time.Time) string {
			if t == nil {
				return ""
			}

			return t.Format(time.RFC3339)
		},
		"format_date": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"format_date_null": func(t *time.Time) string {
			if t == nil {
				return ""
			}

			return t.Format("2006-01-02")
		},
		"format_time_relative": func(t time.Time) string {
			return time.Now().Sub(t).String()
		},
		"pascal_to_snake": stringutil.PascalToSnake,
		"pascal_to_title": stringutil.PascalToTitle,
		"make_map": func(args ...interface{}) map[string]interface{} {
			m := make(map[string]interface{})

			for i := 0; i < len(args)/2; i++ {
				kv := args[i*2]
				vv := args[i*2+1]

				k, ok := kv.(string)
				if !ok {
					panic(fmt.Errorf("key value should be string; was instead %T", kv))
				}

				m[k] = vv
			}

			return m
		},
		"make_string_list": func(items ...string) []string {
			return items
		},
	}

	var templates templatecollection.Collection

	if directoryExists("templates") {
		l.Info("using live filesystem for templates")
		c, err := templatecollection.NewLive(os.DirFS("templates"), templateFuncs)
		if err != nil {
			return fmt.Errorf("runApplicationWorker: %w", err)
		}
		templates = c
	} else {
		l.Info("using embedded filesystem for templates")
		c, err := templatecollection.NewCached(templateFS, templateFuncs)
		if err != nil {
			return fmt.Errorf("runApplicationWorker: %w", err)
		}
		templates = c
	}

	m := mux.NewRouter()

	m.Methods(http.MethodGet).Path("/").HandlerFunc(handlers.Index)
	m.Methods(http.MethodGet).Path("/add").HandlerFunc(handlers.Add)
	m.Methods(http.MethodPost).Path("/add").HandlerFunc(handlers.AddAction)
	m.Methods(http.MethodGet).Path("/channels").HandlerFunc(handlers.Channels)
	m.Methods(http.MethodGet).Path("/channels/{id}").HandlerFunc(handlers.Channel)
	m.Methods(http.MethodGet).Path("/channels/{id}/audio").HandlerFunc(handlers.ChannelAudio)
	m.Methods(http.MethodGet).Path("/channels/{id}/audio-zip").HandlerFunc(handlers.ChannelAudioZip)
	m.Methods(http.MethodGet).Path("/playlists").HandlerFunc(handlers.Playlists)
	m.Methods(http.MethodGet).Path("/playlists/{id}").HandlerFunc(handlers.Playlist)
	m.Methods(http.MethodGet).Path("/playlists/{id}/audio").HandlerFunc(handlers.PlaylistAudio)
	m.Methods(http.MethodGet).Path("/playlists/{id}/audio-zip").HandlerFunc(handlers.PlaylistAudioZip)
	m.Methods(http.MethodGet).Path("/playlists/{id}/{index}").HandlerFunc(handlers.PlaylistVideo)
	m.Methods(http.MethodGet).Path("/videos").HandlerFunc(handlers.Videos)
	m.Methods(http.MethodGet).Path("/videos/audio").HandlerFunc(handlers.VideosAudio)
	m.Methods(http.MethodGet).Path("/videos/audio-zip").HandlerFunc(handlers.VideosAudioZip)
	m.Methods(http.MethodGet).Path("/videos/{id}").HandlerFunc(handlers.Video)
	m.Methods(http.MethodGet).Path("/jobs").HandlerFunc(handlers.Jobs)
	m.Methods(http.MethodGet).Path("/jobs/updates").HandlerFunc(handlers.JobsSSE)

	if directoryExists("static") {
		l.Info("using live filesystem for static files")
		m.Methods(http.MethodGet).PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	} else {
		l.Info("using embedded filesystem for static files")
		m.Methods(http.MethodGet).PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	m.Methods(http.MethodGet).PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir(ctxconfig.GetConfig(ctx).ApplicationDataPath))))

	min := minify.New()
	min.Add("text/html", html.DefaultMinifier)
	min.Add("text/css", css.DefaultMinifier)
	min.Add("application/javascript", js.DefaultMinifier)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.UseFunc(ctxlogger.Register(l))
	n.UseFunc(ctxtimer.Register(nil))
	n.UseFunc(ctxclock.Register(ctxclock.GetClock(ctx)))
	n.UseFunc(ctxtemplate.Register(templates))
	n.UseFunc(ctxdb.Register(ctxdb.GetDB(ctx)))
	n.UseFunc(ctxjobqueue.Register(ctxjobqueue.GetWorker(ctx)))
	n.UseFunc(ctxtimer.AddLoggerHooks())
	n.UseFunc(ctxclock.AddLoggerHooks())
	n.UseFunc(ctxlogger.Log())

	n.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(ctxtemplate.WithData(r.Context(), map[string]interface{}{
			"Messages": struct{ Error, Success, Information string }{
				r.URL.Query().Get("error"),
				r.URL.Query().Get("success"),
				r.URL.Query().Get("information"),
			},
		})))
	})

	if cfg.ApplicationMinify {
		n.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if strings.ToLower(r.Header.Get("connection")) != "upgrade" {
				mw := min.ResponseWriter(rw, r)
				defer mw.Close()
				rw = mw
			}

			next(rw, r)
		})
	}

	n.UseHandler(m)

	s := &http.Server{
		Addr:        addr,
		Handler:     n,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}

	errs := make(chan error, 1)
	go func() {
		l.Info("starting server")
		errs <- s.ListenAndServe()
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		return s.Shutdown(ctx)
	}
}

func registerJobQueueWorkerFunctions(ctx context.Context) error {
	l := ctxlogger.GetLogger(ctx)

	l.WithFields(logrus.Fields{}).Info("registering job queue worker functions")

	w := ctxjobqueue.GetWorker(ctx)
	if w == nil {
		return fmt.Errorf("job queue worker not available in context")
	}

	return w.RegisterAll(map[string]jobqueue.WorkerFunction{
		queuenames.ChannelUpdateMetadata: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			externalID, _, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			channelData, err := ytdirect.GetChannel(ctx, externalID)
			if err != nil {
				return "", err
			}

			if err := ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				var channel models.Channel
				if err := sorm.FindFirstWhere(ctx, tx, &channel, "where external_id = ?", externalID); err != nil {
					if err != sql.ErrNoRows {
						return err
					}

					channel.CreatedAt = time.Now()
					channel.ExternalID = externalID
					channel.Title = channelData.Title
					channel.MetadataUpdatedAt = ptr.Time(time.Now())

					return sorm.CreateRecord(ctx, tx, &channel)
				} else {
					channel.Title = channelData.Title
					channel.MetadataUpdatedAt = ptr.Time(time.Now())

					return sorm.SaveRecord(ctx, tx, &channel)
				}
			}); err != nil {
				return "", err
			}

			return "", nil
		},
		queuenames.ChannelUpdatePlaylists: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			id, _, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			var channel models.Channel
			if err := sorm.FindFirstWhere(ctx, ctxdb.GetDB(ctx), &channel, "where id = ?", id); err != nil {
				return "", err
			}

			channelData, err := ytdirect.GetChannel(ctx, channel.ExternalID)
			if err != nil {
				return "", err
			}

			if err := ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				for _, channelShelf := range channelData.Shelves {
					for _, channelPlaylist := range channelShelf.Playlists {
						var playlist models.Playlist
						if err := sorm.FindFirstWhere(ctx, tx, &playlist, "where external_id = ?", channelPlaylist.ID); err != nil {
							if err != sql.ErrNoRows {
								return err
							}

							playlist.ExternalID = channelPlaylist.ID
							playlist.ChannelID = &channel.ID
							playlist.ChannelExternalID = channel.ExternalID
							playlist.Title = channelPlaylist.Title
							playlist.MetadataUpdatedAt = ptr.Time(time.Now())

							if err := sorm.CreateRecord(ctx, tx, &playlist); err != nil {
								return err
							}
						} else {
							playlist.ExternalID = channelPlaylist.ID
							playlist.ChannelID = &channel.ID
							playlist.ChannelExternalID = channel.ExternalID
							playlist.Title = channelPlaylist.Title
							playlist.MetadataUpdatedAt = ptr.Time(time.Now())

							if err := sorm.SaveRecord(ctx, tx, &playlist); err != nil {
								return err
							}
						}
					}
				}

				return nil
			}); err != nil {
				return "", err
			}

			return "", nil
		},
		queuenames.PlaylistUpdateMetadata: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			externalID, _, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			playlistData, err := ytdirect.GetPlaylist(ctx, externalID)
			if err != nil {
				return "", err
			}

			if err := ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				var channelID *int

				if playlistData.ChannelID != "" {
					if err := tx.QueryRowContext(ctx, "select id from channels where external_id = ?", playlistData.ChannelID).Scan(&channelID); err != nil {
						if err != sql.ErrNoRows {
							return err
						}

						if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
							QueueName: queuenames.ChannelUpdateMetadata,
							Payload:   playlistData.ChannelID,
						}); err != nil {
							return err
						}
					}
				}

				var playlist models.Playlist
				if err := sorm.FindFirstWhere(ctx, tx, &playlist, "where external_id = ?", externalID); err != nil {
					if err != sql.ErrNoRows {
						return err
					}

					playlist.CreatedAt = time.Now()
					playlist.ExternalID = externalID
					playlist.ChannelID = channelID
					playlist.ChannelExternalID = playlistData.ChannelID
					playlist.Title = playlistData.Title
					playlist.MetadataUpdatedAt = ptr.Time(time.Now())

					if err := sorm.CreateRecord(ctx, tx, &playlist); err != nil {
						return err
					}

					for i, videoID := range playlistData.VideoIDs {
						if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
							QueueName: queuenames.VideoUpdateMetadata,
							Payload:   videoID,
						}); err != nil {
							return err
						}

						if err := sorm.CreateRecord(ctx, tx, &models.PlaylistVideo{
							CreatedAt:          time.Now(),
							PlaylistID:         playlist.ID,
							PlaylistExternalID: playlist.ExternalID,
							VideoID:            nil,
							VideoExternalID:    videoID,
							Position:           i,
						}); err != nil {
							return err
						}
					}
				} else {
					playlist.ExternalID = externalID
					if channelID != nil {
						playlist.ChannelID = channelID
					}
					if playlistData.ChannelID != "" {
						playlist.ChannelExternalID = playlistData.ChannelID
					}
					if playlistData.Title != "" {
						playlist.Title = playlistData.Title
					}
					playlist.MetadataUpdatedAt = ptr.Time(time.Now())

					if err := sorm.SaveRecord(ctx, tx, &playlist); err != nil {
						return err
					}

					for i, videoID := range playlistData.VideoIDs {
						if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
							QueueName: queuenames.VideoUpdateMetadata,
							Payload:   videoID,
						}); err != nil {
							return err
						}

						var video models.Video
						if err := sorm.FindFirstWhere(ctx, tx, &video, "where external_id = ?", videoID); err != nil && err != sql.ErrNoRows {
							return err
						}

						var playlistVideo models.PlaylistVideo
						if err := sorm.FindFirstWhere(ctx, tx, &playlistVideo, "where playlist_external_id = ? and video_external_id = ?", playlist.ExternalID, videoID); err != nil {
							if err != sql.ErrNoRows {
								return err
							}

							playlistVideo.CreatedAt = time.Now()
							playlistVideo.PlaylistID = playlist.ID
							playlistVideo.PlaylistExternalID = playlist.ExternalID
							if video.ID != 0 {
								playlistVideo.VideoID = &video.ID
							}
							playlistVideo.VideoExternalID = videoID
							playlistVideo.Position = i

							if err := sorm.CreateRecord(ctx, tx, &playlistVideo); err != nil {
								return err
							}
						} else {
							playlistVideo.PlaylistID = playlist.ID
							playlistVideo.PlaylistExternalID = playlist.ExternalID
							if video.ID != 0 {
								playlistVideo.VideoID = &video.ID
							}
							playlistVideo.VideoExternalID = videoID
							playlistVideo.Position = i

							if err := sorm.SaveRecord(ctx, tx, &playlistVideo); err != nil {
								return err
							}
						}
					}
				}

				return nil
			}); err != nil {
				return "", err
			}

			return "", nil
		},
		queuenames.VideoUpdateMetadata: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			externalID, _, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			videoData, err := ytdirect.GetVideo(ctx, externalID)
			if err != nil {
				return "", err
			}

			var publishDate *time.Time
			if t, err := time.Parse("2006-01-02", videoData.PublishDate); err == nil {
				publishDate = &t
			}
			var uploadDate *time.Time
			if t, err := time.Parse("2006-01-02", videoData.UploadDate); err == nil {
				uploadDate = &t
			}

			if err := ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				var channelID *int
				if err := tx.QueryRowContext(ctx, "select id from channels where external_id = ?", videoData.ChannelID).Scan(&channelID); err != nil {
					if err != sql.ErrNoRows {
						return err
					}

					if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
						QueueName: queuenames.ChannelUpdateMetadata,
						Payload:   videoData.ChannelID,
					}); err != nil {
						return err
					}
				}

				var video models.Video
				if err := sorm.FindFirstWhere(ctx, tx, &video, "where external_id = ?", externalID); err != nil {
					if err != sql.ErrNoRows {
						return err
					}

					video.CreatedAt = time.Now()
					video.ExternalID = externalID
					video.ChannelID = channelID
					video.ChannelExternalID = videoData.ChannelID
					video.Title = videoData.Title
					video.Description = videoData.Description
					video.PublishDate = publishDate
					video.UploadDate = uploadDate
					video.MetadataUpdatedAt = ptr.Time(time.Now())

					if err := sorm.CreateRecord(ctx, tx, &video); err != nil {
						return err
					}

					if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
						QueueName: queuenames.VideoDownload,
						Payload:   externalID,
					}); err != nil {
						return err
					}
				} else {
					video.ExternalID = externalID
					video.ChannelID = channelID
					video.ChannelExternalID = videoData.ChannelID
					video.Title = videoData.Title
					video.Description = videoData.Description
					video.PublishDate = publishDate
					video.UploadDate = uploadDate
					video.MetadataUpdatedAt = ptr.Time(time.Now())

					if err := sorm.SaveRecord(ctx, tx, &video); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return "", err
			}

			return "", nil
		},
		queuenames.VideoDownload: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			externalID, _, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			var video models.Video
			if err := sorm.FindFirstWhere(ctx, ctxdb.GetDB(ctx), &video, "where external_id = ?", externalID); err != nil {
				return "", err
			}

			if video.DownloadedAt != nil {
				return "", nil
			}

			if _, err := os.Stat(cfg.DataFile("videos", externalID+".mp4")); err != nil {
				// Create progress callback for real-time updates
				progressCallback := func(progress int) {
					if err := w.UpdateProgress(ctx, j, progress); err != nil {
						ctxlogger.GetLogger(ctx).WithError(err).Warn("failed to update progress")
					}
				}
				
				// Use the new progress-enabled download function
				if err := ytdl.DownloadVideoWithProgress(ctx, externalID, cfg.DataFile("videos", externalID+".mp4"), progressCallback); err != nil {
					return "", err
				}
			}

			return "", ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				var video models.Video
				if err := sorm.FindFirstWhere(ctx, tx, &video, "where external_id = ?", externalID); err != nil {
					return err
				}

				video.DownloadedAt = ptr.Time(time.Now())

				if err := sorm.SaveRecord(ctx, tx, &video); err != nil {
					return err
				}

				if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
					QueueName: queuenames.VideoUpdateThumbnail,
					Payload:   externalID,
				}); err != nil {
					return err
				}

				for _, size := range []string{} {
					if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
						QueueName: queuenames.VideoTranscode,
						Payload:   externalID + "?size=" + size,
					}); err != nil {
						return err
					}
				}

				if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
					QueueName: queuenames.VideoExtractAudio,
					Payload:   externalID,
				}); err != nil {
					return err
				}

				return nil
			})
		},
		queuenames.VideoUpdateThumbnail: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			externalID, _, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			var video models.Video
			if err := sorm.FindFirstWhere(ctx, ctxdb.GetDB(ctx), &video, "where external_id = ?", externalID); err != nil {
				return "", err
			}

			if video.DownloadedAt == nil {
				return "", fmt.Errorf("video has not been downloaded")
			}

			output, err := ffmpeg.MakeThumbnail(ctx, cfg.DataFile("videos", externalID+".mp4"), cfg.DataFile("thumbnails", externalID+".jpg"))
			if err != nil {
				return output, err
			}

			return output, ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				var video models.Video
				if err := sorm.FindFirstWhere(ctx, tx, &video, "where external_id = ?", externalID); err != nil {
					return err
				}

				video.ThumbnailUpdatedAt = ptr.Time(time.Now())

				if err := sorm.SaveRecord(ctx, tx, &video); err != nil {
					return err
				}

				return nil
			})
		},
		queuenames.VideoTranscode: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			externalID, params, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			size := params.Get("size")
			switch size {
			case "360", "720":
			default:
				return "", fmt.Errorf("transcode size should be 360 or 720")
			}

			var video models.Video
			if err := sorm.FindFirstWhere(ctx, ctxdb.GetDB(ctx), &video, "where external_id = ?", externalID); err != nil {
				return "", err
			}

			if video.DownloadedAt == nil {
				return "", fmt.Errorf("video has not been downloaded")
			}

			var output string

			if _, err := os.Stat(cfg.DataFile("videos", externalID+"_"+size+".mp4")); err != nil {
				// Create progress callback for real-time updates
				progressCallback := func(progress int) {
					if err := w.UpdateProgress(ctx, j, progress); err != nil {
						ctxlogger.GetLogger(ctx).WithError(err).Warn("failed to update progress")
					}
				}
				
				// Use the new progress-enabled transcode function
				s, err := ffmpeg.TranscodeWithProgress(ctx, cfg.DataFile("videos", externalID+".mp4"), size+":-2", cfg.DataFile("videos", externalID+"_"+size+".mp4"), progressCallback)
				if err != nil {
					return s, err
				}

				output = s
			}

			return output, ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				var video models.Video
				if err := sorm.FindFirstWhere(ctx, tx, &video, "where external_id = ?", externalID); err != nil {
					return err
				}

				switch size {
				case "360":
					video.Transcoded360At = ptr.Time(time.Now())
				case "720":
					video.Transcoded720At = ptr.Time(time.Now())
				default:
					return fmt.Errorf("transcode size should be 360 or 720")
				}

				if err := sorm.SaveRecord(ctx, tx, &video); err != nil {
					return err
				}

				return nil
			})
		},
		queuenames.VideoExtractAudio: func(ctx context.Context, w *jobqueue.Worker, j *jobqueue.Job) (string, error) {
			externalID, _, err := jobqueue.ParsePayload(j.Payload)
			if err != nil {
				return "", err
			}

			var video models.Video
			if err := sorm.FindFirstWhere(ctx, ctxdb.GetDB(ctx), &video, "where external_id = ?", externalID); err != nil {
				return "", err
			}

			if video.DownloadedAt == nil {
				return "", fmt.Errorf("video has not been downloaded")
			}

			var output string

			if _, err := os.Stat(cfg.DataFile("audio", externalID+".mp3")); err != nil {
				s, err := ffmpeg.ExtractAudio(ctx, cfg.DataFile("videos", externalID+".mp4"), cfg.DataFile("audio", externalID+".mp3"))
				if err != nil {
					return s, err
				}
				output = s
			}

			return output, ctxdb.UsingTx(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				var video models.Video
				if err := sorm.FindFirstWhere(ctx, tx, &video, "where external_id = ?", externalID); err != nil {
					return err
				}

				video.AudioExtractedAt = ptr.Time(time.Now())

				if err := sorm.SaveRecord(ctx, tx, &video); err != nil {
					return err
				}

				return nil
			})
		},
	})
}

func runJobQueueWorker(ctx context.Context) error {
	l := ctxlogger.GetLogger(ctx)

	l.WithFields(logrus.Fields{}).Info("running job queue worker")

	w := ctxjobqueue.GetWorker(ctx)
	if w == nil {
		return fmt.Errorf("job queue worker not available in context")
	}

	return w.Run(ctx)
}
