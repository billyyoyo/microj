package logger

import (
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/util"
	"github.com/billyyoyo/viper"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
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
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
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

func Info(msg ...any) {
	_log.Info().Msg(fmt.Sprint(msg...))
}

func Infof(format string, msg ...any) {
	_log.Info().Msgf(format, msg...)
}

func Warn(msg ...any) {
	_log.Warn().Msg(fmt.Sprint(msg...))
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
	return json.MarshalIndent(v, "", "\t")
}
