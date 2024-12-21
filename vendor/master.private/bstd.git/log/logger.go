package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Logger struct {
	level  byte
	writer io.Writer
}

type LoggerCfgFunc func(*Logger) error

var ErrInvalidLogLevel = errors.New("log: invalid log level")

const (
	logTrace byte = 6
	logDebug byte = 5
	logInfo  byte = 4
	logWarn  byte = 3
	logErr   byte = 2
	logFatal byte = 1
)

func NewLogger(cfgFunc ...LoggerCfgFunc) (*Logger, error) {
	var (
		err    error
		logger = Logger{logInfo, os.Stderr}
	)
	for _, fn := range cfgFunc {
		err = fn(&logger)
		if err != nil {
			return nil, err
		}
	}
	return &logger, nil
}

func WithWriter(w io.Writer) LoggerCfgFunc {
	return func(l *Logger) error {
		l.writer = w
		return nil
	}
}

func WithLevel(level string) LoggerCfgFunc {
	return func(l *Logger) error {
		l.level = logStringToByte(level)
		return nil
	}
}

func (l *Logger) Trace(a ...interface{}) {
	var msg string
	if !l.toLog(logTrace) {
		return
	}
	msg = fmt.Sprintln(a...)
	l.log(logTrace, msg)
}

func (l *Logger) Tracef(format string, a ...interface{}) {
	var msg string
	if !l.toLog(logTrace) {
		return
	}
	msg = fmt.Sprintf(format+"\n", a...)
	l.log(logTrace, msg)
}

func (l *Logger) Debug(a ...interface{}) {
	var msg string
	if !l.toLog(logDebug) {
		return
	}
	msg = fmt.Sprintln(a...)
	l.log(logDebug, msg)
}

func (l *Logger) Debugf(format string, a ...interface{}) {
	var msg string
	if !l.toLog(logDebug) {
		return
	}
	msg = fmt.Sprintf(format+"\n", a...)
	l.log(logDebug, msg)
}

func (l *Logger) Info(a ...interface{}) {
	var msg string
	if !l.toLog(logInfo) {
		return
	}
	msg = fmt.Sprintln(a...)
	l.log(logInfo, msg)
}

func (l *Logger) Infof(format string, a ...interface{}) {
	var msg string
	if !l.toLog(logInfo) {
		return
	}
	msg = fmt.Sprintf(format+"\n", a...)
	l.log(logInfo, msg)
}

func (l *Logger) Warn(a ...interface{}) {
	var msg string
	if !l.toLog(logWarn) {
		return
	}
	msg = fmt.Sprintln(a...)
	l.log(logWarn, msg)
}

func (l *Logger) Warnf(format string, a ...interface{}) {
	var msg string
	if !l.toLog(logWarn) {
		return
	}
	msg = fmt.Sprintf(format+"\n", a...)
	l.log(logWarn, msg)
}

func (l *Logger) Err(a ...interface{}) {
	var msg string
	if !l.toLog(logErr) {
		return
	}
	msg = fmt.Sprintln(a...)
	l.log(logErr, msg)
}

func (l *Logger) Errf(format string, a ...interface{}) {
	var msg string
	if !l.toLog(logErr) {
		return
	}
	msg = fmt.Sprintf(format+"\n", a...)
	l.log(logErr, msg)
}

func (l *Logger) Fatal(a ...interface{}) {
	var msg string
	if !l.toLog(logFatal) {
		return
	}
	msg = fmt.Sprintln(a...)
	l.log(logFatal, msg)
}

func (l *Logger) Fatalf(format string, a ...interface{}) {
	var msg string
	if !l.toLog(logFatal) {
		return
	}
	msg = fmt.Sprintf(format+"\n", a...)
	l.log(logFatal, msg)
}

func (l *Logger) log(level byte, message string) {
	var (
		msg string
	)
	msg = fmt.Sprintf(
		"%s |%s| %s",
		time.Now().Format(time.DateTime),
		logLevelToStr(level),
		message,
	)
	l.writer.Write([]byte(msg))
	if level == logFatal {
		panic(msg)
	}
}

func (l *Logger) toLog(lvl byte) bool {
	return lvl <= l.level
}

func logLevelToStr(level byte) string {
	switch level {
	case logTrace:
		return "TRACE"
	case logDebug:
		return "DEBUG"
	case logInfo:
		return "INFO"
	case logWarn:
		return "WARN"
	case logErr:
		return "ERR"
	case logFatal:
		return "FATAL"
	default:
		return ""
	}
}

func logStringToByte(level string) byte {
	switch strings.ToUpper(level) {
	case "TRACE":
		return logTrace
	case "DEBUG":
		return logDebug
	case "INFO":
		return logInfo
	case "WARN":
		return logWarn
	case "ERR":
		return logErr
	case "FATAL":
		return logFatal
	default:
		return logInfo
	}
}
