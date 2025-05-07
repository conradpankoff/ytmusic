package ctxdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"fknsrs.biz/p/ytmusic/internal/dbsavepoint"
)

var (
	ErrNoDB = fmt.Errorf("ctxdb: no db found in context")
)

// context registration

var dbKey int

func WithDB(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, &dbKey, db)
}

func GetDB(ctx context.Context) *sql.DB {
	if v := ctx.Value(&dbKey); v != nil {
		return v.(*sql.DB)
	}

	return nil
}

var savepointKey int

func WithSavepoint(ctx context.Context, sp *dbsavepoint.Savepoint) context.Context {
	return context.WithValue(ctx, &savepointKey, sp)
}

func GetSavepoint(ctx context.Context) *dbsavepoint.Savepoint {
	if v := ctx.Value(&savepointKey); v != nil {
		return v.(*dbsavepoint.Savepoint)
	}

	return nil
}

func createSavepoint(ctx context.Context, name string) (context.Context, *dbsavepoint.Savepoint, error) {
	if parent := GetSavepoint(ctx); parent != nil {
		sp, err := dbsavepoint.CreateFromParent(ctx, parent, name)
		if err != nil {
			return ctx, nil, err
		}

		return WithSavepoint(ctx, sp), sp, nil
	}

	if db := GetDB(ctx); db != nil {
		sp, err := dbsavepoint.CreateFromDB(ctx, db, name)
		if err != nil {
			return ctx, nil, err
		}

		return WithSavepoint(ctx, sp), sp, nil
	}

	return ctx, nil, ErrNoDB
}

type SavepointFunc func(ctx context.Context, sp *dbsavepoint.Savepoint) error

func UsingSavepoint(ctx context.Context, name string, fn SavepointFunc) error {
	ctx2, sp, err := createSavepoint(ctx, name)
	if err != nil {
		return err
	}

	if err := fn(ctx2, sp); err != nil {
		if err2 := sp.Rollback(ctx); err2 != nil {
			return errors.Join(err, err2)
		}

		return err
	}

	if err := sp.Release(ctx); err != nil {
		return err
	}

	return nil
}

type TxFunc func(ctx context.Context, tx *sql.Tx) error

func UsingTx(ctx context.Context, opts *sql.TxOptions, fn TxFunc) error {
	db := GetDB(ctx)
	if db == nil {
		return ErrNoDB
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// middleware

func Register(db *sql.DB) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithDB(r.Context(), db)))
	}
}
