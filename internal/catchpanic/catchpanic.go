package catchpanic

import (
	"fmt"
)

func Catch(fn func()) (err error) {
	defer func() {
		if ex := recover(); ex != nil {
			if err1, ok := ex.(error); ok {
				err = fmt.Errorf("catchpanic.Catch: %w", err1)
			} else {
				err = fmt.Errorf("catchpanic.Catch: %s", ex)
			}
		}
	}()

	fn()

	return
}

func CatchErr0(fn func() error) (error) {
	var err error

	if err1 := Catch(func() { err = fn() }); err1 != nil {
		return err1
	}

	return err
}

func CatchErr1[T any](fn func() (T, error)) (T, error) {
	var res T
	var err error

	if err1 := Catch(func() { res, err = fn() }); err1 != nil {
		err = err1
	}

	return res, err
}
