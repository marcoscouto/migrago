package errors

import "errors"

var (
	ErrDuplicatedFile = errors.New("duplicated migration file")
	ErrOutOfOrder     = errors.New("migration out of order")
	ErrInvalidPattern = errors.New("invalid migration filename format")
)
