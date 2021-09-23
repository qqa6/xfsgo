package log

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"time"
)
const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)
type Formatter struct {

}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	//msg := entry.Message
	//for k, v := range entry.Data {
	//	data[k] = v
	//}
	buf := bytes.NewBuffer(nil)
	_,_ = fmt.Fprintf(buf, "%s ", entry.Time.Format(time.RFC3339))
	printLogLevel(buf, entry.Level, false)
	_,_ = fmt.Fprintf(buf, "%s ", firstUpper(entry.Message))
	printFields(buf, entry.Data)
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
func firstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func printFields(w io.Writer, fields logrus.Fields) {
	i := 0
	for k, v := range fields {
		suffix := ""
		if i < len(fields) - 1 {
			suffix = ", "
		}
		_,_ = fmt.Fprintf(w,"%s=%s%s", k, v,suffix)
		i+=1
	}
}

func printLogLevel(w io.Writer, level logrus.Level, color bool) {
	var levelColor int
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	case logrus.InfoLevel:
		levelColor = blue
	default:
		levelColor = blue
	}
	levelString := strings.ToUpper(level.String())
	if color {
		_, _ = fmt.Fprintf(w,"\x1b[%dm%s\x1b[0m ",levelColor, levelString)
	}else {
		_, _ = fmt.Fprintf(w,"%s ", levelString)
	}
}