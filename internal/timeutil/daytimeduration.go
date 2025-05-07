package timeutil

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"time"
)

func ParseDayTimeDuration(s string) (DayTimeDuration, error) {
	var d DayTimeDuration
	if err := d.UnmarshalText([]byte(s)); err != nil {
		return 0, fmt.Errorf("timeutil.ParseDayTimeDuration: %w", err)
	}

	return d, nil
}

type dayTimeDurationSegment struct {
	c byte
	d time.Duration
	t bool
}

var dayTimeDurationSegments = []dayTimeDurationSegment{
	{'D', time.Hour * 24, false},
	{'H', time.Hour, true},
	{'M', time.Minute, true},
	{'S', time.Second, true},
}

type DayTimeDuration int64

func (d *DayTimeDuration) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("timeutil.DayTimeDuration.UnmarshalText: empty duration string")
	}

	var sign int64 = 1
	if b[0] == '-' {
		sign = -1
		b = b[1:]
	}

	if b[0] != 'P' {
		return fmt.Errorf("timeutil.DayTimeDuration.UnmarshalText: could not find 'P' separator")
	}
	b = b[1:]

	total := int64(0)
	rest := string(b)
	readT := false

	for _, e := range dayTimeDurationSegments {
		if len(rest) == 0 {
			break
		}

		if rest[0] == 'T' && !e.t {
			continue
		}

		if e.t && !readT {
			if rest[0] != 'T' {
				return fmt.Errorf("timeutil.DayTimeDuration.UnmarshalText: could not find 'T' separator")
			}
			rest = rest[1:]

			readT = true
		}

		var i int
		var sawDecimal bool
	loop:
		for i = 0; i < len(rest); i++ {
			switch {
			case rest[i] >= '0' && rest[i] <= '9':
				// continue
			case rest[i] == '.' && sawDecimal == false:
				sawDecimal = true
				// continue
			default:
				break loop
			}
		}

		if rest[i] != e.c {
			continue
		}

		f, err := strconv.ParseFloat(rest[:i], 64)
		if err != nil {
			return fmt.Errorf("timeutil.DayTimeDuration.UnmarshalText: couldn't parse segment '%c': %w", e.c, err)
		}
		rest = rest[i+1:]

		if math.Remainder(f, 1) != 0 && e.c != 'S' {
			return fmt.Errorf("timeutil.DayTimeDuration.UnmarshalText: segment '%c' can not have a fractional component", e.c)
		}

		total = total + int64(f*float64(e.d))
	}

	if len(rest) != 0 {
		return fmt.Errorf("timeutil.DayTimeDuration.UnmarshalText: leftover data after parsing is complete: %q", rest)
	}

	*d = DayTimeDuration(sign * total)

	return nil
}

func (d *DayTimeDuration) MarshalText() ([]byte, error) {
	var b bytes.Buffer

	dd := int64(*d)

	if dd < 0 {
		dd = 0 - dd

		if _, err := b.WriteString("-"); err != nil {
			return nil, fmt.Errorf("timeutil.DayTimeDuration.MarshalText: could not write negative sign: %w", err)
		}
	}

	if _, err := b.WriteString("P"); err != nil {
		return nil, fmt.Errorf("timeutil.DayTimeDuration.MarshalText: could not write 'P' separator: %w", err)
	}

	wroteT := false

	for _, e := range dayTimeDurationSegments {
		if e.c == 'S' {
			if e.t && !wroteT {
				wroteT = true
				if _, err := b.WriteString("T"); err != nil {
					return nil, fmt.Errorf("timeutil.DayTimeDuration.MarshalText: could not write 'T' separator: %w", err)
				}
			}

			if _, err := b.WriteString(strconv.FormatFloat(float64(dd)/float64(time.Second), 'f', -1, 64) + "S"); err != nil {
				return nil, fmt.Errorf("timeutil.DayTimeDuration.MarshalText: could not write '%c' component: %w", e.c, err)
			}
		} else if v := dd / int64(e.d); v > 0 {
			dd -= v * int64(e.d)

			if e.t && !wroteT {
				wroteT = true
				if _, err := b.WriteString("T"); err != nil {
					return nil, fmt.Errorf("timeutil.DayTimeDuration.MarshalText: could not write 'T' separator: %w", err)
				}
			}

			if _, err := b.WriteString(strconv.Itoa(int(v)) + string(e.c)); err != nil {
				return nil, fmt.Errorf("timeutil.DayTimeDuration.MarshalText: could not write '%c' component: %w", e.c, err)
			}
		}
	}

	return b.Bytes(), nil
}
