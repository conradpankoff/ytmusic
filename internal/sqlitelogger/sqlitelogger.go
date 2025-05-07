package sqlitelogger

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	proxy "github.com/shogo82148/go-sql-proxy"
	"github.com/sirupsen/logrus"

	"fknsrs.biz/p/ytmusic/internal/ctxclock"
	"fknsrs.biz/p/ytmusic/internal/ctxlogger"
	"fknsrs.biz/p/ytmusic/internal/stackutil"
)

var (
	ErrCancelLogging = fmt.Errorf("cancel logging")
)

type Stats struct {
	Start    time.Time
	Duration time.Duration
	Stack    []runtime.Frame

	query     string
	queryText string
	queryArgs []driver.NamedValue
}

func (s *Stats) Query() string {
	if s.query == "" && s.queryText != "" {
		s.query = printQuery(s.queryText, s.queryArgs)
	}
	return s.query
}

type Filter interface {
	PreCollection(ctx context.Context, stats *Stats) error
	PreLogging(ctx context.Context, stats *Stats) error
	HideStackFrame(ctx context.Context, index int, frame runtime.Frame) (bool, error)
}

func makeStats(ctx context.Context, stmt *proxy.Stmt, args []driver.NamedValue, filters []Filter) (*Stats, error) {
	now, err := ctxclock.Now(ctx)
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		Start: now,
		Stack: stackutil.GetStack(100, 1),
	}

	if stmt != nil {
		stats.queryText = stmt.QueryString
		stats.queryArgs = args
	}

	for _, filter := range filters {
		if err := filter.PreCollection(ctx, stats); err != nil {
			if err == ErrCancelLogging {
				return nil, nil
			}

			return nil, err
		}
	}

	return stats, nil
}

func logStats(ctx context.Context, upperError error, qctx interface{}, filters []Filter, prefix, message string) error {
	if upperError != nil {
		return upperError
	}

	if qctx == nil {
		return nil
	}

	stats, ok := qctx.(*Stats)
	if !ok {
		return nil
	}

	if stats == nil {
		return nil
	}

	now, err := ctxclock.Now(ctx)
	if err != nil {
		return err
	}

	stats.Duration = now.Sub(stats.Start)

	for _, filter := range filters {
		if err := filter.PreLogging(ctx, stats); err != nil {
			if err == ErrCancelLogging {
				return nil
			}

			return err
		}
	}

	fields := logrus.Fields{
		prefix + ".start":    stats.Start.Format(time.RFC3339),
		prefix + ".duration": stats.Duration,
		prefix + ".content":  stats.Query(),
	}

loop:
	for index, frame := range stats.Stack {
		for _, filter := range filters {
			hide, err := filter.HideStackFrame(ctx, index, frame)
			if err != nil {
				return err
			}
			if hide {
				continue loop
			}
		}

		fields[fmt.Sprintf("%s.stack.%02d", prefix, index)] = stackutil.FormatStackFrame(frame)
	}

	ctxlogger.GetLogger(ctx).WithFields(fields).Info(message)

	return nil
}

func New(name string, wrapped driver.Driver, filters ...Filter) driver.Driver {
	return proxy.NewProxyContext(wrapped, &proxy.HooksContext{
		PrePrepare: func(ctx context.Context, stmt *proxy.Stmt) (interface{}, error) {
			return makeStats(ctx, stmt, nil, filters)
		},
		PostPrepare: func(ctx context.Context, qctx interface{}, stmt *proxy.Stmt, err error) error {
			return logStats(ctx, err, qctx, filters, "sql.prepare", "sql prepare")
		},
		PreExec: func(ctx context.Context, stmt *proxy.Stmt, args []driver.NamedValue) (interface{}, error) {
			return makeStats(ctx, stmt, args, filters)
		},
		PostExec: func(ctx context.Context, qctx interface{}, stmt *proxy.Stmt, args []driver.NamedValue, _ driver.Result, err error) error {
			return logStats(ctx, err, qctx, filters, "sql.exec", "sql exec")
		},
		PreQuery: func(ctx context.Context, stmt *proxy.Stmt, args []driver.NamedValue) (interface{}, error) {
			return makeStats(ctx, stmt, args, filters)
		},
		PostQuery: func(ctx context.Context, qctx interface{}, stmt *proxy.Stmt, args []driver.NamedValue, _ driver.Rows, err error) error {
			return logStats(ctx, err, qctx, filters, "sql.query", "sql query")
		},
		PreBegin: func(ctx context.Context, conn *proxy.Conn) (interface{}, error) {
			return makeStats(ctx, nil, nil, filters)
		},
		PostBegin: func(ctx context.Context, qctx interface{}, conn *proxy.Conn, err error) error {
			return logStats(ctx, err, qctx, filters, "sql.tx_begin", "sql tx begin")
		},
		PreCommit: func(ctx context.Context, tx *proxy.Tx) (interface{}, error) {
			return makeStats(ctx, nil, nil, filters)
		},
		PostCommit: func(ctx context.Context, qctx interface{}, tx *proxy.Tx, err error) error {
			return logStats(ctx, err, qctx, filters, "sql.tx_commit", "sql tx commit")
		},
		PreRollback: func(ctx context.Context, tx *proxy.Tx) (interface{}, error) {
			return makeStats(ctx, nil, nil, filters)
		},
		PostRollback: func(ctx context.Context, qctx interface{}, tx *proxy.Tx, err error) error {
			return logStats(ctx, err, qctx, filters, "sql.tx_rollback", "sql tx rollback")
		},
	})
}

type BasicFilter struct {
	CancelAll                bool
	LogSlowerThan            time.Duration
	IgnorePackageStackFrames []string
	IgnoreFunctionQueries    []string
	PreCollectionFunc        func(ctx context.Context, stats *Stats) error
	PreLoggingFunc           func(ctx context.Context, stats *Stats) error
}

func (b *BasicFilter) PreCollection(ctx context.Context, stats *Stats) error {
	if b.CancelAll {
		return ErrCancelLogging
	}

	for _, functionName := range b.IgnoreFunctionQueries {
		for _, frame := range stats.Stack {
			if frame.Function == functionName {
				return ErrCancelLogging
			}
		}
	}

	if b.PreCollectionFunc != nil {
		if err := b.PreCollectionFunc(ctx, stats); err != nil {
			return err
		}
	}

	return nil
}

func (b *BasicFilter) PreLogging(ctx context.Context, stats *Stats) error {
	if b.CancelAll {
		return ErrCancelLogging
	}

	if b.LogSlowerThan != 0 && stats.Duration < b.LogSlowerThan {
		return ErrCancelLogging
	}

	if b.PreLoggingFunc != nil {
		if err := b.PreLoggingFunc(ctx, stats); err != nil {
			return err
		}
	}

	return nil
}

func (b *BasicFilter) HideStackFrame(ctx context.Context, index int, frame runtime.Frame) (bool, error) {
	for _, packageName := range b.IgnorePackageStackFrames {
		if strings.HasPrefix(frame.Function, packageName+".") {
			return true, nil
		}
	}

	return false, nil
}

func printQuery(sqlString string, args []driver.NamedValue) string {
	re := regexp.MustCompile(`\$([0-9]+)`)
	ws := regexp.MustCompile(`\s+`)

	const sqlNull = "NULL"

	return strings.TrimSpace(ws.ReplaceAllString(re.ReplaceAllStringFunc(sqlString, func(s string) string {
		i, err := strconv.ParseInt(s[1:], 10, 64)
		if err != nil {
			return s
		}

		if i < 1 || int(i) > len(args) {
			return s
		}

		switch e := args[i-1].Value.(type) {
		case bool:
			return fmt.Sprintf("%t", e)
		case *bool:
			if e == nil {
				return sqlNull
			}

			return fmt.Sprintf("%t", *e)
		case sql.NullBool:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("%v", e.Bool)
		case *sql.NullBool:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("%v", e.Bool)
		case float64:
			return fmt.Sprintf("%f", e)
		case *float64:
			if e == nil {
				return sqlNull
			}

			return fmt.Sprintf("%f", *e)
		case sql.NullFloat64:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("%f", e.Float64)
		case *sql.NullFloat64:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("%f", e.Float64)
		case int:
			return fmt.Sprintf("%d", e)
		case *int:
			if e == nil {
				return sqlNull
			}

			return fmt.Sprintf("%d", *e)
		case int64:
			return fmt.Sprintf("%d", e)
		case *int64:
			if e == nil {
				return sqlNull
			}

			return fmt.Sprintf("%d", *e)
		case sql.NullInt64:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("%v", e.Int64)
		case *sql.NullInt64:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("%v", e.Int64)
		case string:
			return fmt.Sprintf("'%v'", e)
		case *string:
			if e == nil {
				return sqlNull
			}

			return fmt.Sprintf("'%v'", *e)
		case sql.NullString:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("'%v'", e.String)
		case *sql.NullString:
			if !e.Valid {
				return sqlNull
			}

			return fmt.Sprintf("'%v'", e.String)
		case time.Time:
			return fmt.Sprintf("'%s'", e.Format(time.RFC3339Nano))
		case *time.Time:
			if e == nil {
				return sqlNull
			}

			return fmt.Sprintf("'%s'", e.Format(time.RFC3339Nano))
		case []byte:
			s := fmt.Sprintf("%s", e)
			if r, ok := printable(s); !ok {
				return fmt.Sprintf("[%d bytes of binary data (%q)]", len(s), r)
			}

			return fmt.Sprintf("'%s'", s)
		default:
			s := fmt.Sprintf("%v", args[i-1].Value)
			if r, ok := printable(s); !ok {
				return fmt.Sprintf("[%d bytes of binary data (%q)]", len(s), r)
			}

			return fmt.Sprintf("'%s'", s)
		}
	}), " "))
}

func makePrintable(s string) string {
	if r, ok := printable(s); !ok {
		return fmt.Sprintf("[%d bytes of binary data (%q)]", len(s), r)
	}
	return s
}

func printable(s string) (rune, bool) {
	for _, r := range s {
		if unicode.IsControl(r) {
			return r, false
		}

		if unicode.IsPrint(r) {
			continue
		}

		if r > unicode.MaxASCII {
			return r, false
		}
	}

	return 0, true
}
