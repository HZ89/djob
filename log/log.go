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

package log

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	//	"github.com/rifflock/lfshook"
)

// FmdLoger store logger
var FmdLoger = logrus.NewEntry(logrus.New())

// InitLogger used to  initialization logger
func InitLogger(logLevel string, node string, file string) {
	formattedLogger := logrus.New()
	if file != "" {
		fd, err := os.Open(file)
		if err != nil && err != os.ErrNotExist {
			FmdLoger.WithError(err).Fatal("Open log file failed")
		}
		if fd == nil {
			fd, err = os.Create(file)
			if err != nil {
				FmdLoger.WithError(err).Fatal("Create log file failed")
			}
		}
		fd.Close()
		writer, err := rotatelogs.New(
			file+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(file),
			rotatelogs.WithRotationTime(time.Duration(86400)*time.Second),
		)
		if err != nil {
			FmdLoger.WithError(err).Fatal("Create rotate log failed")
		}
		formattedLogger.Out = writer
	}

	formattedLogger.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithError(err).Error("Error parsing log level, using: info")
		level = logrus.InfoLevel
	}

	formattedLogger.Level = level
	FmdLoger = logrus.NewEntry(formattedLogger).WithField("node", node)
}
