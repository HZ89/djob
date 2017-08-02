package scheduler

import "errors"

var (
	ErrParseDate    = errors.New("Failed to parse date")
	ErrUnrecongDesc = errors.New("Unrecognized descriptor")
)
