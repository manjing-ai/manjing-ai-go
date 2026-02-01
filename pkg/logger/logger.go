package logger

import "github.com/sirupsen/logrus"

var log = logrus.New()

// L returns the global logger instance.
func L() *logrus.Logger {
	return log
}

// Init sets basic logger configuration.
func Init() {
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
