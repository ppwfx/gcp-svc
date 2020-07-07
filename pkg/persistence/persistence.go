package persistence

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ppwfx/user-svc/pkg/types"
	"time"
)

func WaitForDb(connection string) (err error) {
	// simulate trying to connect to database
	time.Sleep(3 * time.Second)

	return
}

func GetDb(maxOpen int, maxIdle int, maxLifetime time.Duration, connection string) (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("postgres", connection)
	if err != nil {
		return
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(maxLifetime)

	return
}

func Migrate(db *sqlx.DB) (err error) {
	_, err = db.Exec(`CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(256) UNIQUE NOT NULL,
		password VARCHAR(256) NOT NULL,
		fullname VARCHAR(256) NOT NULL
	)`)
	if err != nil {
		return
	}

	return
}

func InsertUser(db *sqlx.DB, u types.UserModel) (err error) {
	_, err = db.NamedExec("INSERT INTO users (email, password, fullname) VALUES (:email, :password, :fullname)", &u)
	if err != nil {
		return
	}

	return
}

func SelectUsersOrderByIdDesc(db *sqlx.DB) (us []types.UserModel, err error) {
	err = db.Select(&us, "SELECT id, email, fullname FROM users ORDER BY id DESC")
	if err != nil {
		return
	}

	return
}

func GetUserByEmail(db *sqlx.DB, e string) (u types.UserModel, err error) {
	err = db.Get(&u, "SELECT id, email, fullname, password FROM users WHERE email=$1", e)
	if err != nil {
		return
	}

	return
}

func DeleteUserByEmail(db *sqlx.DB, e string) (err error) {
	_, err = db.Exec("DELETE FROM users WHERE email=$1", e)
	if err != nil {
		return
	}

	return
}
