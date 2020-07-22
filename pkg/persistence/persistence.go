package persistence

import (
	"context"
	"github.com/ppwfx/user-svc/pkg/utils"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ppwfx/user-svc/pkg/types"
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

func Migrate(ctx context.Context, db *sqlx.DB) (err error) {
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogTook, time.Since(begin).String(),
			types.LogSec, time.Since(begin).Seconds(),
			types.LogFunc, "Migrate",
		)

		if err != nil {
			l.Error(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	_, err = db.ExecContext(ctx, `CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(256) UNIQUE NOT NULL,
		password VARCHAR(256) NOT NULL,
		role VARCHAR(256) NOT NULL,
		fullname VARCHAR(256) NOT NULL
	)`)
	if err != nil {
		return
	}

	return
}

func InsertUser(ctx context.Context, db *sqlx.DB, u types.UserModel) (err error) {
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogTook, time.Since(begin).String(),
			types.LogSec, time.Since(begin).Seconds(),
			types.LogFunc, "InsertUser",
			"user_role", u.Role,
		)

		if err != nil {
			l.Error(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	_, err = db.NamedExecContext(ctx, "INSERT INTO users (email, password, fullname, role) VALUES (:email, :password, :fullname, :role)", &u)
	if err != nil {
		return
	}

	return
}

func SelectUsersOrderByIdDesc(ctx context.Context, db *sqlx.DB) (us []types.UserModel, err error) {
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogTook, time.Since(begin).String(),
			types.LogSec, time.Since(begin).Seconds(),
			types.LogFunc, "SelectUsersOrderByIdDesc",
			"returned_users_count", len(us),
		)

		if err != nil {
			l.Error(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	err = db.SelectContext(ctx, &us, "SELECT id, email, fullname FROM users ORDER BY id DESC")
	if err != nil {
		return
	}

	return
}

func GetUserByEmail(ctx context.Context, db *sqlx.DB, e string) (u types.UserModel, err error) {
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogTook, time.Since(begin).String(),
			types.LogSec, time.Since(begin).Seconds(),
			types.LogFunc, "GetUserByEmail",
			"returned_user_id", u.ID,
		)

		if err != nil {
			l.Error(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	err = db.GetContext(ctx, &u, "SELECT id, email, fullname, role, password FROM users WHERE email=$1", e)
	if err != nil {
		return
	}

	return
}

func DeleteUserByEmail(ctx context.Context, db *sqlx.DB, e string) (err error) {
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogTook, time.Since(begin).String(),
			types.LogSec, time.Since(begin).Seconds(),
			types.LogFunc, "DeleteUserByEmail",
		)

		if err != nil {
			l.Error(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	_, err = db.ExecContext(ctx, "DELETE FROM users WHERE email=$1", e)
	if err != nil {
		return
	}

	return
}
