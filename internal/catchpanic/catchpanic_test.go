package catchpanic

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatchError(t *testing.T) {
	a := assert.New(t)

	err := Catch(func() { panic(fmt.Errorf("test_error")) })
	a.Error(err)
	a.ErrorContains(err, "test_error")
}

func TestCatchString(t *testing.T) {
	a := assert.New(t)

	err := Catch(func() { panic("test_error") })
	a.Error(err)
	a.ErrorContains(err, "test_error")
}

func TestCatchErr0(t *testing.T) {
	a := assert.New(t)

	{
		err := CatchErr0(func() error { return fmt.Errorf("test_error") })
		a.Error(err)
		a.ErrorContains(err, "test_error")
	}

	{
		err := CatchErr0(func() error { panic(fmt.Errorf("test_error")) })
		a.Error(err)
		a.ErrorContains(err, "test_error")
	}

	{
		err := CatchErr0(func() error { panic("test_error") })
		a.Error(err)
		a.ErrorContains(err, "test_error")
	}
}

func TestCatchErr1(t *testing.T) {
	a := assert.New(t)

	{
		v, err := CatchErr1(func() (string, error) { return "test_result", nil })
		a.Equal(v, "test_result")
		a.NoError(err)
	}

	{
		v, err := CatchErr1(func() (string, error) { return "test_result", fmt.Errorf("test_error") })
		a.Equal(v, "test_result")
		a.Error(err)
		a.ErrorContains(err, "test_error")
	}

	{
		v, err := CatchErr1(func() (string, error) { return "", fmt.Errorf("test_error") })
		a.Equal(v, "")
		a.Error(err)
		a.ErrorContains(err, "test_error")
	}

	{
		v, err := CatchErr1(func() (string, error) { panic(fmt.Errorf("test_error")); return "", nil })
		a.Equal(v, "")
		a.Error(err)
		a.ErrorContains(err, "test_error")
	}

	{
		v, err := CatchErr1(func() (string, error) { panic("test_error"); return "", nil })
		a.Equal(v, "")
		a.Error(err)
		a.ErrorContains(err, "test_error")
	}
}
