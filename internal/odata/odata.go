package odata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Value interface {
	Type() string
	String() string
}

type PrimitiveValue struct {
	typeName    string
	boolValue   bool
	floatValue  float64
	intValue    int64
	stringValue string
	timeValue   *time.Time
}

var (
	uuidRegexp   = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	singleRegexp = regexp.MustCompile(`^-?[0-9]+.[0-9]+f$`)
	doubleRegexp = regexp.MustCompile(`^-?[0-9]+((.[0-9]+)|[E[+|-][0-9]+])d$`)
	int32Regexp  = regexp.MustCompile(`^-?[0-9]+$`)
	int64Regexp  = regexp.MustCompile(`^-?[0-9]+L$`)
	timeRegexp   = regexp.MustCompile(`^-?P([0-9]+D)?T([0-9]+H)?([0-9]+M)?([0-9]+(\.[0-9]+)?S)?$`)
)

func ParsePrimitiveValue(input string) (*PrimitiveValue, error) {
	// Null
	if input == "null" {
		return &PrimitiveValue{typeName: "Null"}, nil
	}

	// Edm.Boolean
	if input == "true" || input == "false" {
		value, err := strconv.ParseBool(input)
		if err != nil {
			return nil, err
		}
		return &PrimitiveValue{typeName: "Edm.Boolean", boolValue: value}, nil
	}

	// Edm.DateTime
	if strings.HasPrefix(input, "datetime'") && strings.HasSuffix(input, "'") {
		value, err := time.Parse(time.RFC3339, input[9:len(input)-1])
		if err != nil {
			return nil, err
		}

		return &PrimitiveValue{typeName: "Edm.DateTime", timeValue: &value}, nil
	}

	// Edm.Guid
	if strings.HasPrefix(input, "guid'") && strings.HasSuffix(input, "'") {
		value := input[5 : len(input)-1]
		if !uuidRegexp.MatchString(value) {
			return nil, fmt.Errorf("odata.ParsePrimitiveValue: guid format was incorrect")
		}

		return &PrimitiveValue{typeName: "Edm.Guid", stringValue: value}, nil
	}

	// Edm.Time
	if timeRegexp.MatchString(input) {
		value, err := timeutil.ParseDayTimeDuration(input)
		if err != nil {
			return nil, fmt.Errorf("odata.ParsePrimitiveValue: could not parse as time (duration) value: %w", err)
		}

		return &PrimitiveValue{typeName: "Edm.Time", timeValue: &value}, nil
	}

	// Edm.Single
	// Example 1: 2.0f
	if singleRegexp.MatchString(input) {
		value, err := strconv.ParseFloat(input[0:len(input)-1], 32)
		if err != nil {
			return nil, fmt.Errorf("odata.ParsePrimitiveValue: could not parse as single-length floating point value: %w", err)
		}
		return &PrimitiveValue{typeName: "Edm.Single", floatValue: float64(singleValue)}, nil
	}

	// Edm.Double
	// Example 1: 1E+10d
	// Example 2: 2.029d
	// Example 3: 2.0d
	if doubleRegexp.MatchString(input) {
		value, err := strconv.ParseFloat(input[0:len(input)-1], 64)
		if err != nil {
			return nil, fmt.Errorf("odata.ParsePrimitiveValue: could not parse as double-length floating point value: %w", err)
		}

		return &PrimitiveValue{typeName: "Edm.Double", floatValue: value}, nil
	}

	// Edm.Int32
	if int32Regexp.MatchString(input) {
		value, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("odata.ParsePrimitiveValue: could not parse as 32-bit integer value: %w", err)
		}
		return &PrimitiveValue{typeName: "Edm.Int32", intValue: value}, nil
	}

	// Edm.Int64
	if int64Regexp.MatchString(input) {
		value, err := strconv.ParseInt(input[0:len(input)-1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("odata.ParsePrimitiveValue: could not parse as 64-bit integer value: %w", err)
		}
		return &PrimitiveValue{typeName: "Edm.Int64", intValue: value}, nil
	}

	// Edm.String
	return &PrimitiveValue{typeName: "Edm.String", stringValue: input}, nil
}

func (p *PrimitiveValue) Type() string {
	return p.typeName
}

func (p *PrimitiveValue) Bool() bool {
	return
}
