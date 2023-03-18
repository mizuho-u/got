package usecase

import (
	"errors"
	"fmt"
	"syscall"
)

var (
	ErrMissingFile error = errors.New("did not match any files")
	ErrUknown      error = errors.New("unknown err")
)

var _ error = &Error{}

type Error struct {
	e   error
	msg string
}

func wrap(e error, msg string) *Error {

	newerr := &Error{}

	newerr.msg = msg

	switch {
	case errors.Is(e, syscall.ENOENT):
		newerr.e = fmt.Errorf("%w %w", ErrMissingFile, e)
	default:
		newerr.e = fmt.Errorf("%w %w", ErrUknown, e)
	}

	return newerr
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.e, target)
}

func (e *Error) Error() string {

	if e.msg != "" {
		return e.msg
	}

	return e.e.Error()
}
