package communication

import (
	"encoding/json"
	"github.com/armon/go-metrics"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/types"
	"go.uber.org/zap"
)

func Serve(v *validator.Validate, logger *zap.SugaredLogger, m *metrics.Metrics, db *sqlx.DB, addr string, hmacSecret string, allowedSubjectSuffix string, argon2IdOpts business.Argon2IdOpts) (err error) {
	mux := http.NewServeMux()

	var maxBodyBytes int64 = 256 * 1024

	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return composeContextLoggerMiddleware(logger,
			secureMiddleware(
				composeMaxBodyBytesMiddleware(maxBodyBytes,
					composeAuthMiddleware(hmacSecret,
						authorizationMiddleware(next),
					),
				),
			),
		)
	}

	defaultMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return composeContextLoggerMiddleware(logger,
			secureMiddleware(
				composeMaxBodyBytesMiddleware(maxBodyBytes,
					authorizationMiddleware(next),
				),
			),
		)
	}

	mux.HandleFunc(types.RouteCreateUser, defaultMiddleware(func(w http.ResponseWriter, r *http.Request) {
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

		rsp, statusCode = business.CreateUser(r.Context(), m, db, argon2IdOpts, v, allowedSubjectSuffix, req)

		return
	}))

	mux.HandleFunc(types.RouteListUsers, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			statusCode = http.StatusNotAcceptable

			return
		}

		rsp, statusCode = business.ListUsers(r.Context(), m, db, v, req)

		return
	}))

	mux.HandleFunc(types.RouteDeleteUser, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var rsp types.DeleteUserResponse
		var statusCode int

		defer func() {
			err := r.Body.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close request body")

				logger.Error(err)
			}

			writeJsonResponse(logger, w, statusCode, rsp)
		}()

		var req types.DeleteUserRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			statusCode = http.StatusNotAcceptable

			return
		}

		if strings.HasSuffix(req.Email, allowedSubjectSuffix) {
			rsp.Error = types.ErrorCanNotDeleteInternalUser
			statusCode = http.StatusUnprocessableEntity
		}

		rsp, statusCode = business.DeleteUser(r.Context(), m, db, v, req)

		return
	}))

	mux.HandleFunc(types.RouteAuthenticate, sensitiveMiddleware(defaultMiddleware(func(w http.ResponseWriter, r *http.Request) {
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

		rsp, statusCode = business.Authenticate(r.Context(), m, db, v, hmacSecret, req)

		return
	})))

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	logger.Infof("listening on %v\n", addr)

	err = http.Serve(l, mux)
	if err != nil {
		return
	}

	return
}

func extractAccessToken(r *http.Request) (t string) {
	return strings.TrimPrefix(r.Header.Get(types.HeaderAuthorization), types.PrefixBearer)
}

func writeJsonResponse(logger *zap.SugaredLogger, w http.ResponseWriter, statusCode int, rsp interface{}) {
	b, err := json.Marshal(rsp)
	if err != nil {
		logger.Error(err)

		w.WriteHeader(statusCode)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err = w.Write(b)
	if err != nil {
		logger.Error(err)
	}
}
