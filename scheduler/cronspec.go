package scheduler

import "time"

type cronspec struct {
	Second, Minute, Hour, Dom, Month, Dow uint64
}

type section struct {
	max, min uint
	names    map[string]uint
}

var (
	seconds = section{0, 59, nil}
	minutes = section{0, 59, nil}
	hours   = section{0, 23, nil}
	dom     = section{1, 31, nil}
	months  = section{1, 12, map[string]uint{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	dow = section{0, 6, map[string]uint{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)

func (s * section) Next(t time.Time) time.Time{
	
}
