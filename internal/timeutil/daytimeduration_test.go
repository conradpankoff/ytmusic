package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var dayTimeDurationUnmarshalTextTests = []struct {
	name  string
	input string
	value DayTimeDuration
	error string
}{
	{
		name:  "simple duration",
		input: "PT10S",
		value: DayTimeDuration(time.Second * 10),
		error: "",
	},
	{
		name:  "negative duration",
		input: "-PT10S",
		value: DayTimeDuration(time.Second * -10),
		error: "",
	},
	{
		name:  "duration with multiple components",
		input: "P1DT2H3M4S",
		value: DayTimeDuration(time.Hour*24 + time.Hour*2 + time.Minute*3 + time.Second*4),
		error: "",
	},
	{
		name:  "duration with decimal seconds",
		input: "PT1.234S",
		value: DayTimeDuration(time.Second*1 + time.Millisecond*234),
		error: "",
	},
	{
		name:  "invalid duration",
		input: "ABC",
		value: DayTimeDuration(0),
		error: "timeutil.DayTimeDuration.UnmarshalText: could not find 'P' separator",
	},
}

func TestDayTimeDurationUnmarshalText(t *testing.T) {
	for _, tc := range dayTimeDurationUnmarshalTextTests {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			var d DayTimeDuration

			err := d.UnmarshalText([]byte(tc.input))

			if tc.error != "" {
				a.EqualError(err, tc.error)
			} else {
				a.NoError(err)
			}

			a.Equal(tc.value, d)
		})
	}
}

func BenchmarkDayTimeDurationUnmarshalText(b *testing.B) {
	for _, tc := range dayTimeDurationUnmarshalTextTests {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var d DayTimeDuration
				d.UnmarshalText([]byte(tc.input))
			}
		})
	}
}

var dayTimeDurationMarshalTextTests = []struct {
	name  string
	input DayTimeDuration
	value string
	error string
}{
	{
		name:  "simple duration",
		input: DayTimeDuration(time.Second * 10),
		value: "PT10S",
		error: "",
	},
	{
		name:  "negative duration",
		input: DayTimeDuration(time.Second * -10),
		value: "-PT10S",
		error: "",
	},
	{
		name:  "duration with multiple components",
		input: DayTimeDuration(time.Hour*24 + time.Hour*2 + time.Minute*3 + time.Second*4),
		value: "P1DT2H3M4S",
		error: "",
	},
	{
		name:  "duration with decimal seconds",
		input: DayTimeDuration(time.Second + time.Millisecond*500),
		value: "PT1.5S",
		error: "",
	},
}

func TestDayTimeDurationMarshalText(t *testing.T) {
	for _, tc := range dayTimeDurationMarshalTextTests {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			value, err := tc.input.MarshalText()

			if tc.error != "" {
				a.EqualError(err, tc.error)
			} else {
				a.NoError(err)
			}

			a.Equal(tc.value, string(value))
		})
	}
}

func BenchmarkDayTimeDurationMarshalText(b *testing.B) {
	for _, tc := range dayTimeDurationMarshalTextTests {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tc.input.MarshalText()
			}
		})
	}
}
