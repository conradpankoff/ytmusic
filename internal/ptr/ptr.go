package ptr

import (
	"database/sql"
	"time"
)

func Bool(v bool) *bool                { return &v }
func String(v string) *string          { return &v }
func StringSlice(v []string) *[]string { return &v }
func Int(v int) *int                   { return &v }
func Float64(v float64) *float64       { return &v }
func Time(v time.Time) *time.Time      { return &v }

// SQL null types helpers
func NullTime(v time.Time) sql.NullTime      { return sql.NullTime{Time: v, Valid: true} }
func NullTimeFromPtr(v *time.Time) sql.NullTime { 
	if v == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *v, Valid: true} 
}
func NullInt32(v int32) sql.NullInt32        { return sql.NullInt32{Int32: v, Valid: true} }
func NullInt32FromInt(v int) sql.NullInt32   { return sql.NullInt32{Int32: int32(v), Valid: true} }
func NullInt32FromIntPtr(v *int) sql.NullInt32 { 
	if v == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(*v), Valid: true} 
}
