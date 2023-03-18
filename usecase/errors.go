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

func wrap(e error) error {

	switch {
	case errors.Is(e, syscall.ENOENT):
		return fmt.Errorf("%w %w", ErrMissingFile, e)
	default:
		return fmt.Errorf("%w %w", ErrUknown, e)
	}

}
