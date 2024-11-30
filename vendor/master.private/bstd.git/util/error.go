package util

import (
	"errors"
)

func ErrWrap(message string, err error) error {
	var (
		newErr = errors.New(message)
	)
	return errors.Join(newErr, err)
}
