package djob

import (
	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	"time"
	"github.com/rifflock/lfshook"
)

var Log = logrus.NewEntry(logrus.New())

func InitLogger(logLevel string, node string, file string) {

	formattedLogger := logrus.New()
	formattedLogger.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	if file != ""{
		writer := rotatelogs.New(
			file,
			rotatelogs.WithLinkName(file),
			rotatelogs.WithRotationTime(time.Duration(2073600) * time.Second),
		)
		formattedLogger.Hooks.Add(lfshook.NewHook(lfshook.WriterMap{
			logrus.InfoLevel: writer,
			logrus.DebugLevel: writer,
			logrus.PanicLevel: writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.WarnLevel: writer,
		}))
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithError(err).Error("Error parsing log level, using: info")
		level = logrus.InfoLevel
	}

	formattedLogger.Level = level
	Log = logrus.NewEntry(formattedLogger).WithField("node", node)

}
