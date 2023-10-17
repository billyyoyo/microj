package logger

import (
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/util"
	"github.com/billyyoyo/viper"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
	"time"
)

type LogConf struct {
	Level    string            `yaml:"level"`
	ErrStack bool              `yaml:"errStack"`
	Debug    bool              `yaml:"debug"`
	Caller   bool              `yaml:"caller"`
	File     lumberjack.Logger `yaml:"file"`
}

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorFormat = "\x1b[%dm%v\x1b[0m"
)

var (
	_conf LogConf
	_log  zerolog.Logger
)

type Val struct {
	K string
	V any
}

func init() {
	var err error
	viper.SetConfigType("yaml")
	viper.AddConfigPath(util.RunningSpace() + "conf")
	viper.AddConfigPath(util.RunningSpace())
	viper.SetConfigName("log.yml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Err(err).Msg("log config file load failed")
		return
	}
	err = viper.UnmarshalKey("log", &_conf)
	if err != nil {
		log.Err(err).Msg("log config parse failed")
		return
	}
	// config logger
	lvl, err := zerolog.ParseLevel(_conf.Level)
	if err == nil {
		zerolog.SetGlobalLevel(lvl)
	}
	if _conf.ErrStack {
		//zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.ErrorStackMarshaler = marshalStack
	}
	ctx := zerolog.New(os.Stderr).With().Timestamp().Stack()
	if _conf.Caller {
		ctx = ctx.Caller()
	}
	_log = ctx.Logger()
	writer := zerolog.ConsoleWriter{
		TimeFormat: time.DateTime + ".000",
	}
	if !_conf.Debug {
		//_log = _log.Output(&_conf.File)
		writer.FormatLevel = formatLevelFile()
		writer.NoColor = true
		writer.Out = &_conf.File
		_log = _log.Output(writer)
	} else {
		zerolog.InterfaceMarshalFunc = stackMarshal
		writer.FormatLevel = formatLevelConsole()
		writer.Out = os.Stderr
		_log = _log.Output(writer)
	}
}

func Debug(msg ...any) {
	_log.Debug().Msg(fmt.Sprint(msg...))
}

func Debugf(format string, msg ...any) {
	_log.Debug().Msgf(format, msg...)
}

func Info(msg ...any) {
	_log.Info().Msg(fmt.Sprint(msg...))
}

func Infof(format string, msg ...any) {
	_log.Info().Msgf(format, msg...)
}

func Warn(msg ...any) {
	_log.Warn().Msg(fmt.Sprint(msg...))
}

func Warnf(format string, msg ...any) {
	_log.Warn().Msgf(format, msg...)
}

func Err(err error) {
	ev := _log.Error().Stack()
	if err != nil {
		ev = ev.Err(err)
	}
	ev.Send()
}

func Error(msg string, err error, vals ...Val) {
	ev := _log.Error().Stack()
	if err != nil {
		ev = ev.Err(err)
	}
	if len(vals) > 0 {
		for _, v := range vals {
			ev = ev.Any(v.K, v.V)
		}
	}
	ev.Msg(msg)
}

func Errorf(format string, err error, vars ...any) {
	ev := _log.Error().Stack()
	if err != nil {
		ev = ev.Err(err)
	}
	msg := fmt.Sprintf(format, vars...)
	ev.Msg(msg)
}

func Fatal(msg string, err error) {
	ev := _log.Fatal()
	if err != nil {
		ev = ev.Err(err)
	}
	ev.Stack().Msg(msg)
}

func Panic(msg string, err error, vals ...Val) {
	ev := _log.Panic()
	if err != nil {
		ev = ev.Err(err)
	}
	if len(vals) > 0 {
		for _, v := range vals {
			ev = ev.Any(v.K, v.V)
		}
	}
	ev.Stack().Msg(msg)
}

func formatLevelFile() zerolog.Formatter {
	return func(i interface{}) string {
		var l string
		if ll, ok := i.(string); ok {
			switch ll {
			case zerolog.LevelTraceValue:
				l = "TRACE"
			case zerolog.LevelDebugValue:
				l = "DEBUG"
			case zerolog.LevelInfoValue:
				l = "INFO"
			case zerolog.LevelWarnValue:
				l = "WARN"
			case zerolog.LevelErrorValue:
				l = "ERROR"
			case zerolog.LevelFatalValue:
				l = "FATAL"
			case zerolog.LevelPanicValue:
				l = "PANIC"
			default:
				l = ll
			}
		} else {
			if i == nil {
				l = "???"
			} else {
				l = strings.ToUpper(fmt.Sprintf("%s", i))[0:3]
			}
		}
		return l
	}
}
func formatLevelConsole() zerolog.Formatter {
	return func(i interface{}) string {
		var l string
		if ll, ok := i.(string); ok {
			switch ll {
			case zerolog.LevelTraceValue:
				l = colorize("TRACE", colorWhite)
			case zerolog.LevelDebugValue:
				l = colorize("DEBUG", colorCyan)
			case zerolog.LevelInfoValue:
				l = colorize("INFO", colorGreen)
			case zerolog.LevelWarnValue:
				l = colorize("WARN", colorBlue)
			case zerolog.LevelErrorValue:
				l = colorize("ERROR", colorRed)
			case zerolog.LevelFatalValue:
				l = colorize("FATAL", colorMagenta)
			case zerolog.LevelPanicValue:
				l = colorize("PANIC", colorYellow)
			default:
				l = ll
			}
		} else {
			if i == nil {
				l = "???"
			} else {
				l = strings.ToUpper(fmt.Sprintf("%s", i))[0:3]
			}
		}
		return l
	}
}
func colorize(s interface{}, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func stackMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

type state struct {
	b []byte
}

// Write implement fmt.Formatter interface.
func (s *state) Write(b []byte) (n int, err error) {
	s.b = b
	return len(b), nil
}

// Width implement fmt.Formatter interface.
func (s *state) Width() (wid int, ok bool) {
	return 0, false
}

// Precision implement fmt.Formatter interface.
func (s *state) Precision() (prec int, ok bool) {
	return 0, false
}

// Flag implement fmt.Formatter interface.
func (s *state) Flag(c int) bool {
	return false
}

func frameField(f errors.Frame, s *state, c rune) string {
	f.Format(s, c)
	return string(s.b)
}

func marshalStack(err error) interface{} {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	sterr, ok := err.(stackTracer)
	if !ok {
		return nil
	}
	st := sterr.StackTrace()
	s := &state{}
	out := make([]string, 0, len(st))
	for _, frame := range st {
		out = append(out, fmt.Sprintf("%s -- %s() :: %s", frameField(frame, s, 's'), frameField(frame, s, 'n'), frameField(frame, s, 'd')))
	}
	return out
}
