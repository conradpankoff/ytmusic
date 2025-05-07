package dbsavepoint

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

type querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

var (
	ErrAlreadyRolledBack = fmt.Errorf("dbsavepoint: savepoint already rolled back")
	ErrAlreadyReleased   = fmt.Errorf("dbsavepoint: savepoint already released")
)

type Savepoint struct {
	parent     *Savepoint
	tx         *sql.Tx
	name       string
	ownsTx     bool
	released   bool
	rolledBack bool
}

func CreateFromDB(ctx context.Context, db *sql.DB, name string) (*Savepoint, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf("savepoint %s", name)); err != nil {
		tx.Rollback()
		return nil, err
	}

	return &Savepoint{name: name, tx: tx, ownsTx: true}, nil
}

func CreateFromTx(ctx context.Context, tx *sql.Tx, name string) (*Savepoint, error) {
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("savepoint %s", name)); err != nil {
		return nil, err
	}

	return &Savepoint{name: name, tx: tx}, nil
}

func CreateFromParent(ctx context.Context, sp *Savepoint, name string) (*Savepoint, error) {
	if _, err := sp.querier().ExecContext(ctx, fmt.Sprintf("savepoint %s", name)); err != nil {
		return nil, err
	}

	return &Savepoint{name: sp.name + "." + name, tx: sp.tx, parent: sp}, nil
}

func (sp *Savepoint) querier() querier {
	if sp.parent != nil {
		return sp.parent
	}

	return sp.tx
}

func (sp *Savepoint) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case sp.rolledBack:
		return nil, ErrAlreadyRolledBack
	case sp.released:
		return nil, ErrAlreadyReleased
	}

	return sp.querier().QueryContext(ctx, query, args...)
}

func makeRowWithError(err error) *sql.Row {
	var r sql.Row
	v := reflect.ValueOf(r)
	v.FieldByName("err").Set(reflect.ValueOf(err))
	return &r
}

func (sp *Savepoint) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch {
	case sp.rolledBack:
		return makeRowWithError(ErrAlreadyRolledBack)
	case sp.released:
		return makeRowWithError(ErrAlreadyReleased)
	}

	return sp.querier().QueryRowContext(ctx, query, args...)
}

func (sp *Savepoint) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case sp.rolledBack:
		return nil, ErrAlreadyRolledBack
	case sp.released:
		return nil, ErrAlreadyReleased
	}

	return sp.querier().ExecContext(ctx, query, args...)
}

func (sp *Savepoint) Create(ctx context.Context, name string) (*Savepoint, error) {
	switch {
	case sp.rolledBack:
		return nil, ErrAlreadyRolledBack
	case sp.released:
		return nil, ErrAlreadyReleased
	}

	return CreateFromParent(ctx, sp, name)
}

func (sp *Savepoint) Release(ctx context.Context) error {
	switch {
	case sp.rolledBack:
		return ErrAlreadyRolledBack
	case sp.released:
		return ErrAlreadyReleased
	default:
		sp.released = true
	}

	if _, err := sp.querier().ExecContext(ctx, fmt.Sprintf("release savepoint %s", sp.name)); err != nil {
		return err
	}

	if sp.ownsTx && sp.tx != nil {
		if err := sp.tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (sp *Savepoint) Rollback(ctx context.Context) error {
	switch {
	case sp.rolledBack:
		return ErrAlreadyRolledBack
	case sp.released:
		return ErrAlreadyReleased
	default:
		sp.rolledBack = true
	}

	if _, err := sp.querier().ExecContext(ctx, fmt.Sprintf("rollback to savepoint %s", sp.name)); err != nil {
		return err
	}

	if sp.ownsTx && sp.tx != nil {
		if err := sp.tx.Rollback(); err != nil {
			return err
		}
	}

	return nil
}
