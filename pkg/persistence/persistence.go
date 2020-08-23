package persistence

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/armon/go-metrics"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils/ctxutil"
)

func OpenPostgresDB(maxOpen int, maxIdle int, maxLifetime time.Duration, connection string) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("postgres", connection)
	if err != nil {
		return
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(maxLifetime)

	return
}

func ConnectToPostgresDb(ctx context.Context, db *sqlx.DB, timeout time.Duration) (err error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		err = db.PingContext(ctx)
		if err != nil {
			time.Sleep(100 * time.Millisecond)

			continue
		}

		break
	}
	if err != nil {
		return
	}

	return
}

func Migrate(logger *zap.SugaredLogger, sourceUrl string, pgUrl string) (err error) {
	defer func(begin time.Time) {
		l := logger.With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
		)

		if err != nil {
			l.Error(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	m, err := migrate.New(sourceUrl, pgUrl)
	if err != nil {
		err = errors.Wrap(err, "failed to create *migrate.Migrate instance")

		return
	}
	err = m.Up()
	if err != nil {
		err = errors.Wrap(err, "failed to apply migration")

		return
	}

	return
}

func InsertUser(ctx context.Context, m metrics.MetricSink, db *sqlx.DB, u types.UserModel) (err error) {
	defer func(begin time.Time) {
		l := ctxutil.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
			"user_group", u.UserGroup,
		)

		if err != nil {
			l.Warn(err)
		} else {
			l.Debug()
		}

		m.IncrCounterWithLabels([]string{"persistence", "InsertUser"}, 1, []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
		m.AddSampleWithLabels([]string{"persistence", "InsertUser"}, float32(time.Now().Sub(begin).Milliseconds()), []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
	}(time.Now())

	_, err = db.NamedExecContext(ctx, "INSERT INTO users (email, password, fullname, user_group) VALUES (:email, :password, :fullname, :user_group)", &u)
	if err != nil {
		err = errors.Wrap(err, "failed to insert user")

		return
	}

	return
}

func SelectUsersOrderByIdDesc(ctx context.Context, m metrics.MetricSink, db *sqlx.DB) (us []types.UserModel, err error) {
	defer func(begin time.Time) {
		l := ctxutil.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
			"returned_users_count", len(us),
		)

		if err != nil {
			l.Warn(err)
			m.IncrCounter([]string{"persistence", "SelectUsersOrderByIdDesc", "error"}, 1)
		} else {
			l.Debug()
		}

		m.IncrCounterWithLabels([]string{"persistence", "SelectUsersOrderByIdDesc"}, 1, []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
		m.AddSampleWithLabels([]string{"persistence", "SelectUsersOrderByIdDesc"}, float32(time.Now().Sub(begin).Milliseconds()), []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
	}(time.Now())

	err = db.SelectContext(ctx, &us, "SELECT id, email, fullname FROM users ORDER BY id DESC")
	if err != nil {
		err = errors.Wrap(err, "failed to select users")

		return
	}

	return
}

func GetUserByEmail(ctx context.Context, m metrics.MetricSink, db *sqlx.DB, e string) (u types.UserModel, err error) {
	defer func(begin time.Time) {
		l := ctxutil.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
			"returned_user_id", u.ID,
		)

		if err != nil {
			l.Warn(err)
		} else {
			l.Debug()
		}

		m.IncrCounterWithLabels([]string{"persistence", "GetUserByEmail"}, 1, []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
		m.AddSampleWithLabels([]string{"persistence", "GetUserByEmail"}, float32(time.Now().Sub(begin).Milliseconds()), []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
	}(time.Now())

	err = db.GetContext(ctx, &u, "SELECT id, email, fullname, user_group, password, created_at, updated_at FROM users WHERE email=$1", e)
	if err != nil {
		err = errors.Wrap(err, "failed to select user by email")

		return
	}

	return
}

func DeleteUserByEmail(ctx context.Context, m metrics.MetricSink, db *sqlx.DB, e string) (err error) {
	defer func(begin time.Time) {
		l := ctxutil.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
		)

		if err != nil {
			l.Warn(err)
		} else {
			l.Debug()
		}

		m.IncrCounterWithLabels([]string{"persistence", "DeleteUserByEmail"}, 1, []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
		m.AddSampleWithLabels([]string{"persistence", "DeleteUserByEmail"}, float32(time.Now().Sub(begin).Milliseconds()), []metrics.Label{{Name: "success", Value: strconv.FormatBool(err == nil)}})
	}(time.Now())

	_, err = db.ExecContext(ctx, "DELETE FROM users WHERE email=$1", e)
	if err != nil {
		err = errors.Wrap(err, "failed to delete user by email")

		return
	}

	return
}
