/*
 * Copyright (c) 2017.  Harrison Zhu <wcg6121@gmail.com>
 * This file is part of djob <https://github.com/HZ89/djob>.
 *
 * djob is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * djob is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with djob.  If not, see <http://www.gnu.org/licenses/>.
 */

package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// djob errors list
var (
	ErrNil                   = New(95200, "everything is fine")
	ErrNotExist              = New(95201, "object not exist")
	ErrTimeOut               = New(95202, "time out")
	ErrUnknown               = New(95203, "unknown error")
	ErrUnknownType           = New(95204, "unknown object type")
	ErrCannotSetNilValue     = New(95205, "condition not match")
	ErrEntryTooLarge         = New(95206, "entry is too large")
	ErrTxnTooLarge           = New(95207, "transaction is too large")
	ErrCanNotFoundNode       = New(95208, "could not found any node can use")
	ErrSameJob               = New(95209, "this job set himself as his parent")
	ErrNoCmd                 = New(95210, "a job must have a Command")
	ErrNoReg                 = New(95211, "must specify a region")
	ErrNoExp                 = New(95212, "a job must have a Expression")
	ErrScheduleParse         = New(95213, "can't parse job schedule")
	ErrType                  = New(95214, "type error")
	ErrArgs                  = New(95215, "must have name and region")
	ErrLockTimeout           = New(95216, "locking timeout")
	ErrStartSerf             = New(95217, "start serf failed")
	ErrMissKeyFile           = New(95218, "have no key file or cert file path")
	ErrMissCaFile            = New(95219, "have no ca file path")
	ErrUnknownOps            = New(95220, "unknown ops")
	ErrRepetition            = New(95221, "already have same object")
	ErrLinkNum               = New(95222, "count of links must equal to conditions plus one")
	ErrConditionFormat       = New(95223, "error format of conditions")
	ErrNotSupportSymbol      = New(95224, "logic symbol not support")
	ErrMissApiToken          = New(95225, "must have Api tokens")
	ErrRepetionToken         = New(95226, "api token must be unique")
	ErrNotExpectation        = New(95227, "type does not match expectations")
	ErrIllegalCharacter      = New(95228, "'###' is not allow used in key")
	ErrParentNotInSameRegion = New(95229, "parent job must be in same region")
	ErrParentNotExist        = New(95230, "parent job not exist")
	ErrHaveSubJob            = New(95231, "please delete sub-job first")
	ErrCopyToUnaddressable   = New(95232, "copy to value is unaddressable")
	ErrNodeDead              = New(95233, "found node is not alive")
	ErrNodeNoRPC             = New(95234, "found node have no completed rpc config")
	ErrBlankTimeFormat       = New(95235, "Blank time format")
)

// Error custom error type
type Error struct {
	code    int    // error code
	message string // error message
}

// New a new error
func New(code int, text string) *Error {
	return &Error{
		code:    code,
		message: text,
	}
}

// NewFromGRPCErr func transform grcp error to this Error
func NewFromGRPCErr(err error) (*Error, bool) {
	if err == nil {
		return ErrNil, true
	}
	s, ok := status.FromError(err)
	if !ok {
		return nil, false
	}

	return &Error{code: int(s.Code()), message: s.Message()}, true
}

// Error func implement the standard error interface
func (e *Error) Error() string {
	return fmt.Sprintf("ErrCode=%d, ErrMessage=%s", e.code, e.message)
}

// GenGRPCErr used to generate grpc error
func (e *Error) GenGRPCErr() error {
	return status.Errorf(codes.Code(e.code), e.message)
}

// Equal return true when errors code equal
func (e *Error) Equal(err error) bool {
	if err == nil {
		return false
	}
	if te, ok := err.(*Error); ok {
		return te.code == e.code
	}
	return false
}

// Less error code less
func (e *Error) Less(err error) bool {
	if terr, ok := err.(*Error); ok {
		return e.code < terr.code
	}
	return false
}

// NotEqual return true when errors code not equal
func (e *Error) NotEqual(err error) bool {
	return !e.Equal(err)
}

// GenGRPCErr used to generate grpc error from any kind error
func GenGRPCErr(err error) error {
	if terr, ok := err.(*Error); ok {
		return terr.GenGRPCErr()
	}
	return New(100000, err.Error()).GenGRPCErr()
}
