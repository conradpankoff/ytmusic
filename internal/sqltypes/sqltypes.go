package sqltypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const format = "2006-01-02 15:04:05.999999999-07:00"

type TimeScanner struct {
	Value *time.Time
}

func (t *TimeScanner) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		v, err := time.Parse(format, src)
		if err != nil {
			return fmt.Errorf("sqltypes.TimeScanner: could not parse input value %q: %w", src, err)
		}
		*t.Value = v
		return nil
	default:
		return fmt.Errorf("sqltypes.TimeScanner: could not scan input type of %T", src)
	}
}

type TimePointerScanner struct {
	Value **time.Time
}

func (t *TimePointerScanner) Scan(src interface{}) error {
	switch src := src.(type) {
	case nil:
		*t.Value = nil
		return nil
	case string:
		v, err := time.Parse(format, src)
		if err != nil {
			return fmt.Errorf("sqltypes.TimePointerScanner: could not parse input value %q: %w", src, err)
		}
		*t.Value = &v
		return nil
	default:
		return fmt.Errorf("sqltypes.TimePointerScanner: could not scan input type of %T", src)
	}
}

type StringSliceScanner struct {
	Value *[]string
}

func (t *StringSliceScanner) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		var v []string
		if err := json.Unmarshal([]byte(src), &v); err != nil {
			return fmt.Errorf("sqltypes.StringSliceScanner: could not parse input value as JSON: %w", err)
		}
		*t.Value = v
		return nil
	default:
		return fmt.Errorf("sqltypes.StringSliceScanner: could not scan input type of %T", src)
	}
}

type StringSlicePointerScanner struct {
	Value **[]string
}

func (t *StringSlicePointerScanner) Scan(src interface{}) error {
	switch src := src.(type) {
	case nil:
		*t.Value = nil
		return nil
	case []byte:
		var v []string
		if err := json.Unmarshal(src, &v); err != nil {
			return fmt.Errorf("sqltypes.StringSlicePointerScanner: could not parse input value (%T) as JSON: %w", src, err)
		}
		*t.Value = &v
		return nil
	case string:
		var v []string
		if err := json.Unmarshal([]byte(src), &v); err != nil {
			return fmt.Errorf("sqltypes.StringSlicePointerScanner: could not parse input value (%T) as JSON: %w", src, err)
		}
		*t.Value = &v
		return nil
	default:
		return fmt.Errorf("sqltypes.StringSlicePointerScanner: could not scan input type of %T", src)
	}
}

type JSONStringSlice []string

func (s JSONStringSlice) Value() (driver.Value, error) {
	if s == nil || len(s) == 0 {
		return "[]", nil
	}

	return json.Marshal(s)
}

func (s *JSONStringSlice) Scan(src interface{}) error {
	switch src := src.(type) {
	case nil:
		*s = nil
		return nil
	case []byte:
		if err := json.Unmarshal(src, s); err != nil {
			return fmt.Errorf("sqltypes.JSONStringSlice: could not decode input (%T) as JSON: %w", src, err)
		}
		return nil
	case string:
		if err := json.Unmarshal([]byte(src), s); err != nil {
			return fmt.Errorf("sqltypes.JSONStringSlice: could not decode input (%T) as JSON: %w", src, err)
		}
		return nil
	default:
		return fmt.Errorf("sqltypes.JSONStringSlice: could not scan input type of %T", src)
	}
}
