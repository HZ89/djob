package djob

import "errors"

var (
	ErrNotExist          = errors.New("Error: key not exist")
	ErrTimeOut           = errors.New("Time out")
	ErrUnknown           = errors.New("Unknown error")
	ErrCannotSetNilValue = errors.New("Error: Condition not match")
	ErrEntryTooLarge     = errors.New("entry is too large")
	ErrTxnTooLarge       = errors.New("transaction is too large")
	ErrCanNotFoundNode   = errors.New("could not found any node can use")
	ErrSameJob           = errors.New("This job set himself as his parent")
	ErrNoCmd             = errors.New("A job must have a Command")
	ErrNoReg             = errors.New("A job must have a region")
	ErrNoExp             = errors.New("A job must have a Expression")
	ErrScheduleParse     = errors.New("Can't parse job schedule")
)
