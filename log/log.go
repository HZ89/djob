package log

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	//	"github.com/rifflock/lfshook"
)

var Loger = logrus.NewEntry(logrus.New())

func InitLogger(logLevel string, node string, file string) {
	formattedLogger := logrus.New()
	if file != "" {
		fd, err := os.Open(file)
		if err != nil && err != os.ErrNotExist {
			Loger.WithError(err).Fatal("Open log file failed")
		}
		if fd == nil {
			fd, err = os.Create(file)
			if err != nil {
				Loger.WithError(err).Fatal("Create log file failed")
			}
		}
		fd.Close()
		writer, err := rotatelogs.New(
			file+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(file),
			rotatelogs.WithRotationTime(time.Duration(86400)*time.Second),
		)
		if err != nil {
			Loger.WithError(err).Fatal("Create rotate log failed")
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
	Loger = logrus.NewEntry(formattedLogger).WithField("node", node)
}
