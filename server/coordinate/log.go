package coordinate

import (
	"fmt"
	"github.com/billyyoyo/microj/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
	"io"
	"log"
)

func NewLogger() *logMiddle {
	return &logMiddle{}
}

type logMiddle struct {
	fixedData map[string]interface{}
}

func parseArgs(args []interface{}) (m map[string]interface{}) {
	m = make(map[string]interface{})
	if len(args) == 0 || len(args)%2 > 0 {
		return
	}
	for i := 0; i < len(m)/2; i++ {
		k := args[i]
		switch t := k.(type) {
		case string:
			m[t] = args[i+1]
		default:
			m[fmt.Sprintf("%s", t)] = args[i+1]
		}
	}
	return
}

func (l logMiddle) addFixed(args ...interface{}) map[string]interface{} {
	sm := parseArgs(args)
	for k, v := range l.fixedData {
		sm[k] = v
	}
	return sm
}

func (l logMiddle) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Trace:
		l.Trace(msg, args)
	case hclog.Debug:
		l.Debug(msg, args)
	case hclog.Info:
		l.Info(msg, args)
	case hclog.Warn:
		l.Warn(msg, args)
	case hclog.Error:
		l.Error(msg, args)
	}
}

func (l logMiddle) Trace(msg string, args ...interface{}) {
	m := l.addFixed(args)
	logger.TraceM(msg, m)
}

func (l logMiddle) Debug(msg string, args ...interface{}) {
	m := l.addFixed(args)
	logger.DebugM(msg, m)
}

func (l logMiddle) Info(msg string, args ...interface{}) {
	m := l.addFixed(args)
	logger.InfoM(msg, m)
}

func (l logMiddle) Warn(msg string, args ...interface{}) {
	m := l.addFixed(args)
	logger.WarnM(msg, m)
}

func (l logMiddle) Error(msg string, args ...interface{}) {
	m := l.addFixed(args)
	logger.ErrorM(msg, m)
}

func (l logMiddle) IsTrace() bool {
	return zerolog.Level(logger.Level()) == zerolog.TraceLevel
}

func (l logMiddle) IsDebug() bool {
	return zerolog.Level(logger.Level()) == zerolog.DebugLevel
}

func (l logMiddle) IsInfo() bool {
	return zerolog.Level(logger.Level()) == zerolog.InfoLevel
}

func (l logMiddle) IsWarn() bool {
	return zerolog.Level(logger.Level()) == zerolog.WarnLevel
}

func (l logMiddle) IsError() bool {
	return zerolog.Level(logger.Level()) == zerolog.ErrorLevel
}

func (l logMiddle) ImpliedArgs() []interface{} {
	return nil
}

func (l logMiddle) With(args ...interface{}) hclog.Logger {
	m := parseArgs(args)
	sl := logMiddle{
		fixedData: m,
	}
	return sl
}

func (l logMiddle) Name() string {
	return ""
}

func (l logMiddle) Named(name string) hclog.Logger {
	return l
}

func (l logMiddle) ResetNamed(name string) hclog.Logger {
	return l
}

func (l logMiddle) SetLevel(level hclog.Level) {
	// undo
}

func (l logMiddle) GetLevel() hclog.Level {
	lv := logger.Level()
	switch zerolog.Level(lv) {
	case zerolog.DebugLevel:
		return hclog.Debug
	case zerolog.InfoLevel:
		return hclog.Info
	case zerolog.WarnLevel:
		return hclog.Warn
	case zerolog.ErrorLevel:
		return hclog.Error
	case zerolog.TraceLevel:
		return hclog.Trace
	case zerolog.NoLevel:
		return hclog.NoLevel
	case zerolog.Disabled:
		return hclog.Off
	}
	return hclog.Off
}

func (l logMiddle) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return nil
}

func (l logMiddle) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return nil
}
