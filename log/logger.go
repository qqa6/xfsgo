package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Warningf(string, ...interface{})
	Writer() *io.PipeWriter
	Panic(args ...interface{})
	WithFields(fields logrus.Fields) *logrus.Entry

}

func DefaultLogger() Logger {
	logrus.SetLevel(logrus.InfoLevel)
	return logrus.StandardLogger()
}
