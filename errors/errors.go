package errors

import (
	"errors"

	"github.com/docker/libkv/store"
)

var (
	ErrNotExist          = errors.New("object not exist")
	ErrTimeOut           = errors.New("time out")
	ErrUnknown           = errors.New("unknown error")
	ErrUnknownType       = errors.New("unknown object type")
	ErrCannotSetNilValue = errors.New("condition not match")
	ErrEntryTooLarge     = errors.New("entry is too large")
	ErrTxnTooLarge       = errors.New("transaction is too large")
	ErrCanNotFoundNode   = errors.New("could not found any node can use")
	ErrSameJob           = errors.New("this job set himself as his parent")
	ErrNoCmd             = errors.New("a job must have a Command")
	ErrNoReg             = errors.New("a job must have a region")
	ErrNoExp             = errors.New("a job must have a Expression")
	ErrScheduleParse     = errors.New("can't parse job schedule")
	ErrType              = errors.New("type error")
	ErrArgs              = errors.New("name or Region must have one")
	ErrLockTimeout       = errors.New("locking timeout")
	ErrStartSerf         = errors.New("start serf failed")
	ErrMissKeyFile       = errors.New("have no key file or cert file path")
	ErrMissCaFile        = errors.New("have no ca file path")
	ErrNotFound          = store.ErrKeyNotFound
	ErrUnknownOps        = errors.New("unknown ops")
	ErrRepetition        = errors.New("already have same object")
	ErrLinkNum           = errors.New("count of links must equal to conditions plus one")
	ErrConditionFormat   = errors.New("error format of conditions")
	ErrNotSupportSymbol  = errors.New("logic symbol not support")
)
