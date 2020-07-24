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

type LogHttpRequest struct {
	Method             string
	URL                string
	UserAgent          string
	Referrer           string
	RemoteIP           string
	RequestSize        int64
	ResponseSize       int64
	ResponseStatusCode int
	Latency            string
}

func (r *LogHttpRequest) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("method", r.Method)
	e.AddString("url", r.URL)
	e.AddString("userAgent", r.UserAgent)
	e.AddString("referrer", r.Referrer)
	e.AddInt("responseStatusCode", r.ResponseStatusCode)
	e.AddString("remoteIp", r.RemoteIP)
	e.AddInt64("requestSize", r.RequestSize)
	e.AddInt64("responseSize", r.ResponseSize)
	e.AddString("latency", r.Latency)

	return nil
}

type reportLocation struct {
	filePath     string
	lineNumber   int
	functionName string
}

func (l *reportLocation) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("filePath", l.filePath)
	e.AddInt("lineNumber", l.lineNumber)
	e.AddString("functionName", l.functionName)
	return nil
}

func (c *core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	var functionName string
	fn := runtime.FuncForPC(entry.Caller.PC)
	if fn != nil {
		functionName = strings.TrimSuffix(strings.TrimRight(fn.Name(), "0123456789"), ".func")
	}

	fields = append(fields, zap.Object("context.reportLocation", &reportLocation{
		filePath:     entry.Caller.File,
		lineNumber:   entry.Caller.Line,
		functionName: functionName,
	}))

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
	}), zap.Fields(zap.Object("serviceContext", userLabels{
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
