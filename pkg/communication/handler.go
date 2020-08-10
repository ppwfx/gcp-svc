package communication

import (
	"encoding/json"
	"github.com/armon/go-metrics"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func handleDeleteUser(validator *validator.Validate, logger *zap.SugaredLogger, metrics metrics.MetricSink, db *sqlx.DB, allowedSubjectSuffix string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var rsp types.DeleteUserResponse
		var statusCode int

		l := utils.GetContextLogger(r.Context())

		defer func() {
			err := r.Body.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close request body")

				l.Error(err)
			}

			writeJsonResponse(logger, w, statusCode, rsp)
		}()

		var req types.DeleteUserRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err = errors.Wrapf(err, "failed to decode response")
			l.Error(err)

			statusCode = http.StatusNotAcceptable

			return
		}

		if strings.HasSuffix(req.Email, allowedSubjectSuffix) {
			rsp.Error = types.ErrorCanNotDeleteInternalUser
			statusCode = http.StatusUnprocessableEntity
		}

		rsp, statusCode = business.DeleteUser(r.Context(), metrics, db, validator, req)

		return
	}
}

func handleAuthenticate(validator *validator.Validate, logger *zap.SugaredLogger, metrics metrics.MetricSink, db *sqlx.DB, hmacSecret string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var rsp types.AuthenticateResponse
		var statusCode int

		defer func() {
			err := r.Body.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close request body")

				logger.Error(err)
			}

			writeJsonResponse(logger, w, statusCode, rsp)
		}()

		var req types.AuthenticateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			statusCode = http.StatusNotAcceptable

			return
		}

		rsp, statusCode = business.Authenticate(r.Context(), metrics, db, validator, hmacSecret, req)

		return
	}
}

func handleListUsers(validator *validator.Validate, logger *zap.SugaredLogger, metrics metrics.MetricSink, db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var rsp types.ListUsersResponse
		var statusCode int

		defer func() {
			err := r.Body.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close request body")

				logger.Error(err)
			}

			writeJsonResponse(logger, w, statusCode, rsp)
		}()

		var req types.ListUsersRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			statusCode = http.StatusNotAcceptable

			return
		}

		rsp, statusCode = business.ListUsers(r.Context(), metrics, db, validator, req)

		return
	}
}

func handleCreateUser(validator *validator.Validate, logger *zap.SugaredLogger, metrics metrics.MetricSink, db *sqlx.DB, allowedSubjectSuffix string, argon2IdOpts business.Argon2IdOpts) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var rsp types.CreateUserResponse
		var statusCode int

		defer func() {
			err := r.Body.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close request body")

				logger.Error(err)
			}

			writeJsonResponse(logger, w, statusCode, rsp)
		}()

		var req types.CreateUserRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			statusCode = http.StatusNotAcceptable

			return
		}

		rsp, statusCode = business.CreateUser(r.Context(), metrics, db, argon2IdOpts, validator, allowedSubjectSuffix, req)

		return
	}
}
