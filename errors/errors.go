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
	"errors"
)

var (
	ErrNotExist              = errors.New("object not exist")
	ErrTimeOut               = errors.New("time out")
	ErrUnknown               = errors.New("unknown error")
	ErrUnknownType           = errors.New("unknown object type")
	ErrCannotSetNilValue     = errors.New("condition not match")
	ErrEntryTooLarge         = errors.New("entry is too large")
	ErrTxnTooLarge           = errors.New("transaction is too large")
	ErrCanNotFoundNode       = errors.New("could not found any node can use")
	ErrSameJob               = errors.New("this job set himself as his parent")
	ErrNoCmd                 = errors.New("a job must have a Command")
	ErrNoReg                 = errors.New("must specify a region")
	ErrNoExp                 = errors.New("a job must have a Expression")
	ErrScheduleParse         = errors.New("can't parse job schedule")
	ErrType                  = errors.New("type error")
	ErrArgs                  = errors.New("name or Region must have one")
	ErrLockTimeout           = errors.New("locking timeout")
	ErrStartSerf             = errors.New("start serf failed")
	ErrMissKeyFile           = errors.New("have no key file or cert file path")
	ErrMissCaFile            = errors.New("have no ca file path")
	ErrUnknownOps            = errors.New("unknown ops")
	ErrRepetition            = errors.New("already have same object")
	ErrLinkNum               = errors.New("count of links must equal to conditions plus one")
	ErrConditionFormat       = errors.New("error format of conditions")
	ErrNotSupportSymbol      = errors.New("logic symbol not support")
	ErrMissApiToken          = errors.New("must have Api tokens")
	ErrRepetionToken         = errors.New("api token must be unique")
	ErrNotExpectation        = errors.New("type does not match expectations")
	ErrIllegalCharacter      = errors.New("'###' is not allow used in key")
	ErrParentNotInSameRegion = errors.New("parent job must be in same region")
	ErrParentNotExist        = errors.New("parent job not exist")
	ErrHaveSubJob            = errors.New("please delete sub-job first")
	ErrCopyToUnaddressable   = errors.New("copy to value is unaddressable")
)
