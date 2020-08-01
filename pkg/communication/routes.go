package communication

import (
	"encoding/json"
	"fmt"
	"github.com/armon/go-metrics"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

func GetHandler(v *validator.Validate, logger *zap.SugaredLogger, m metrics.MetricSink, db *sqlx.DB, hmacSecret string, allowedSubjectSuffix string, argon2IdOpts business.Argon2IdOpts) http.Handler {
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
		err := json.NewDecoder(r.Body).Decode(&req)
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

	return mux
}

func GetPprofHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}

func ServePprof(logger *zap.SugaredLogger, port string) (err error) {
	h := GetPprofHandler()

	addr := fmt.Sprintf("0.0.0.0:%v", port)

	s := &http.Server{
		Addr:              fmt.Sprintf(addr),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadTimeout:       5 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           h,
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	logger.Infof("pprof listening on %v", addr)

	err = s.Serve(l)
	if err != nil {
		err = errors.Wrap(err, "failed to serve pprof")

		return
	}

	return
}

func Serve(v *validator.Validate, logger *zap.SugaredLogger, m metrics.MetricSink, db *sqlx.DB, addr string, hmacSecret string, allowedSubjectSuffix string, argon2IdOpts business.Argon2IdOpts) (err error) {
	h := GetHandler(v, logger, m, db, hmacSecret, allowedSubjectSuffix, argon2IdOpts)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	s := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadTimeout:       5 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           h,
	}

	logger.Infof("service listening on %v", addr)

	err = s.Serve(l)
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
