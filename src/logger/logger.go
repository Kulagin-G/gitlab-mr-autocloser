package logger

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
)

func SetupLogger(debug bool) *logrus.Logger {
	logLevel := logrus.InfoLevel

	if debug {
		logLevel = logrus.DebugLevel
	}

	log := logrus.Logger{
		Out:   os.Stdout,
		Level: logLevel,
		Formatter: &prefixed.TextFormatter{
			DisableColors:   false,
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceFormatting: true,
		},
	}

	return &log
}
