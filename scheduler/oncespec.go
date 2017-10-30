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

package scheduler

import (
	"time"
)

// SimpleDelaySchedule represents a simple non recurring duration.
type oncespec struct {
	Date time.Time
}

// Just store the given time for this schedule.
func At(date time.Time) oncespec {
	return oncespec{
		Date: date,
	}
}

// Next conforms to the Schedule interface but this kind of jobs
// doesn't need to be run more than once, so it doesn't return a new date but the existing one.
func (schedule oncespec) Next(t time.Time) time.Time {
	// If the date set is after the reference time return it
	// if it's before, return a virtually infinite sleep date
	// so do nothing.
	if schedule.Date.After(t) {
		return schedule.Date
	}
	return time.Time{}
}
