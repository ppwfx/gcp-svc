package business

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils"
)

func CreateUser(ctx context.Context, db *sqlx.DB, argonOpts Argon2IdOpts, v *validator.Validate, allowedSubjectSuffix string, req types.CreateUserRequest) (rsp types.CreateUserResponse, statusCode int) {
	var err error
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
		)

		if err != nil {
			err = errors.Wrap(err, "failed to create user")

			l.Warn(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	statusCode = http.StatusOK

	err = v.Struct(&req)
	if err != nil {
		err = errors.Wrap(err, "failed to validate the request")

		rsp.Error = err.Error()
		statusCode = http.StatusUnprocessableEntity

		return
	}

	salt, err := generateRandomBytes(argonOpts.SaltLength)
	if err != nil {
		err = errors.Wrap(err, "failed to generate random salt")

		statusCode = http.StatusInternalServerError

		return
	}

	var role string
	if strings.HasSuffix(req.Email, allowedSubjectSuffix) {
		role = types.RoleAdmin
	} else {
		role = types.RoleUser
	}

	err = persistence.InsertUser(ctx, db, types.UserModel{
		Email:    req.Email,
		Password: string(hashSecret(salt, req.Password, argonOpts)),
		FullName: req.FullName,
		Role:     role,
	})
	if err != nil {
		err = errors.Wrap(err, "failed to insert user into database")

		rsp.Error = err.Error()
		statusCode = http.StatusUnprocessableEntity

		return
	}

	return
}

func ListUsers(ctx context.Context, db *sqlx.DB, v *validator.Validate, req types.ListUsersRequest) (rsp types.ListUsersResponse, statusCode int) {
	var err error
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
			"rsp_users_count", len(rsp.Users),
		)

		if err != nil {
			err = errors.Wrap(err, "failed to list users")

			l.Warn(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	statusCode = http.StatusOK

	err = v.Struct(&req)
	if err != nil {
		err = errors.Wrap(err, "failed to validate the request")

		rsp.Error = err.Error()
		statusCode = http.StatusUnprocessableEntity

		return
	}

	us, err := persistence.SelectUsersOrderByIdDesc(ctx, db)
	if err != nil {
		err = errors.Wrap(err, "failed to get users from database")

		rsp.Error = types.ErrorInternalError
		statusCode = http.StatusInternalServerError

		return
	}

	for _, u := range us {
		rsp.Users = append(rsp.Users, types.ListUser{
			ID:       u.ID,
			Email:    u.Email,
			FullName: u.FullName,
		})
	}

	return
}

func DeleteUser(ctx context.Context, db *sqlx.DB, v *validator.Validate, req types.DeleteUserRequest) (rsp types.DeleteUserResponse, statusCode int) {
	var err error
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
		)

		if err != nil {
			err = errors.Wrap(err, "failed to delete user")

			l.Warn(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	statusCode = http.StatusOK

	err = v.Struct(&req)
	if err != nil {
		err = errors.Wrap(err, "failed to validate the request")

		rsp.Error = err.Error()
		statusCode = http.StatusUnprocessableEntity

		return
	}

	_, err = persistence.GetUserByEmail(ctx, db, req.Email)
	if err != nil {
		err = errors.Wrap(err, "failed to get user")

		rsp.Error = types.ErrorUserDoesNotExist
		statusCode = http.StatusUnprocessableEntity

		return
	}

	err = persistence.DeleteUserByEmail(ctx, db, req.Email)
	if err != nil {
		err = errors.Wrap(err, "failed to delete user")

		rsp.Error = types.ErrorInternalError
		statusCode = http.StatusInternalServerError

		return
	}

	return
}

func Authenticate(ctx context.Context, db *sqlx.DB, v *validator.Validate, hmacSecret string, req types.AuthenticateRequest) (rsp types.AuthenticateResponse, statusCode int) {
	var err error
	defer func(begin time.Time) {
		l := utils.GetContextLogger(ctx).With(
			types.LogLatency, fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
		)

		if err != nil {
			err = errors.Wrap(err, "failed to authenticate")

			l.Warn(err)
		} else {
			l.Debug()
		}
	}(time.Now())

	statusCode = http.StatusOK

	err = v.Struct(&req)
	if err != nil {
		err = errors.Wrap(err, "failed to validate the request")

		rsp.Error = err.Error()
		statusCode = http.StatusUnprocessableEntity

		return
	}

	u, err := persistence.GetUserByEmail(ctx, db, req.Email)
	if err != nil {
		err = errors.Wrap(err, "failed to get user from database")

		rsp.Error = types.ErrorInvalidCredentials
		statusCode = http.StatusUnprocessableEntity

		return
	}

	match, err := compareSecretAndHash(req.Password, u.Password)
	if err != nil {
		err = errors.Wrap(err, "failed to compare password and hash")

		rsp.Error = types.ErrorInvalidCredentials
		statusCode = http.StatusUnprocessableEntity

		return
	}
	if !match {
		err = errors.New("failed as password and hash don't match")

		rsp.Error = types.ErrorInvalidCredentials
		statusCode = http.StatusUnprocessableEntity

		return
	}

	accessToken, err := GenerateAccessToken(hmacSecret, u.Role, u.ID)
	if err != nil {
		err = errors.Wrap(err, "failed to generate access token")

		rsp.Error = types.ErrorInternalError
		statusCode = http.StatusInternalServerError

		return
	}

	rsp.AccessToken = accessToken

	return
}
