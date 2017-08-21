package djob

import "errors"

var (
	ErrNotExist          = errors.New("Error: key not exist")
	ErrTimeOut           = errors.New("Time out")
	ErrUnknown           = errors.New("Unknown error")
	ErrCannotSetNilValue = errors.New("Error: Condition not match")
	ErrEntryTooLarge     = errors.New("entry is too large")
	ErrTxnTooLarge       = errors.New("transaction is too large")
)
