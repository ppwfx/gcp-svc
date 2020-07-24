package utils

import (
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var EncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "eventTime",
	LevelKey:       "severity",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "message",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    EncodeLevel,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

var logLevelSeverity = map[zapcore.Level]string{
	zapcore.DebugLevel:  "DEBUG",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARNING",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "CRITICAL",
	zapcore.PanicLevel:  "ALERT",
	zapcore.FatalLevel:  "EMERGENCY",
}

func EncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(logLevelSeverity[lv])
}

type userLabels struct {
	service string
	version string
}

func (s userLabels) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("service", s.service)
	e.AddString("version", s.version)
	return nil
}

type core struct {
	zapcore.Core
}

func (c *core) With(fields []zapcore.Field) zapcore.Core {
	return &core{
		Core: c.Core.With(fields),
	}
}

func (c *core) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}

	return ce
}

func (c *core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	fields = append(fields, zap.String("file", entry.Caller.File))
	fields = append(fields, zap.Int("line", entry.Caller.Line))

	fn := runtime.FuncForPC(entry.Caller.PC)
	if fn != nil {
		fields = append(fields, zap.String("function", fn.Name()))
	}

	strings.SplitN(entry.Caller.File, "pkg/", 1)

	return c.Core.Write(entry, fields)
}

func NewProductionLogger(service string, version string) (sl *zap.SugaredLogger, err error) {
	c := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Encoding:          "json",
		EncoderConfig:     EncoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
	}

	l, err := c.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return &core{Core: c}
	}), zap.Fields(zap.Object("userLabels", userLabels{
		service: service,
		version: version,
	})))
	if err != nil {
		err = errors.Wrap(err, "failed to build zap logger")

		return
	}

	sl = l.Sugar()

	return
}
