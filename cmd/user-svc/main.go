package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/communication"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var args = types.ServeArgs{}

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

func main() {
	flag.StringVar(&args.Addr, "addr", "", "")
	flag.StringVar(&args.DbConnection, "db-connection", "", "")
	flag.StringVar(&args.HmacSecret, "hmac-secret", "", "")
	flag.StringVar(&args.AllowedSubjectSuffix, "allowed-subject-suffix", "", "")
	flag.Parse()

	c := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Encoding:          "json",
		EncoderConfig:     EncoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
	}

	v := validator.New()

	err := func() (err error) {
		l, err := c.Build()
		if err != nil {
			err = errors.Wrap(err, "failed to build zap logger")

			return
		}

		defer func() {
			err := l.Sync()
			if err != nil {
				err = errors.Wrap(err, "failed to flush logger buffer")

				log.Print(err)

				return
			}
		}()

		db, err := persistence.OpenPostgresDB(25, 25, 5*time.Minute, args.DbConnection)
		if err != nil {
			err = errors.Wrap(err, "failed to open postgres")

			return
		}

		err = persistence.ConnectToPostgresDb(context.Background(), db, 5*time.Second)
		if err != nil {
			err = errors.Wrap(err, "failed to connect to postgres")

			return
		}

		err = communication.Serve(v, l.Sugar(), db, args.Addr, args.HmacSecret, args.AllowedSubjectSuffix, business.DefaultArgon2IdOpts)
		if err != nil {
			err = errors.Wrap(err, "failed to listen")

			return
		}

		return
	}()
	if err != nil {
		err = errors.Wrap(err, "failed to run service")

		return
	}

	return
}
