package communication

import (
	"encoding/json"
	"github.com/armon/go-metrics"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/types"
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
	"strings"
)

func AddSvcRoutes(mux *http.ServeMux, validate *validator.Validate, logger *zap.SugaredLogger, metrics metrics.MetricSink, db *sqlx.DB, hmacSecret string, allowedSubjectSuffix string, argon2IdOpts business.Argon2IdOpts) *http.ServeMux {
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

	mux.HandleFunc(types.RouteCreateUser, defaultMiddleware(handleCreateUser(validate, logger, metrics, db, allowedSubjectSuffix, argon2IdOpts)))

	mux.HandleFunc(types.RouteListUsers, authMiddleware(handleListUsers(validate, logger, metrics, db)))

	mux.HandleFunc(types.RouteDeleteUser, authMiddleware(handleDeleteUser(validate, logger, metrics, db, allowedSubjectSuffix)))

	mux.HandleFunc(types.RouteAuthenticate, sensitiveMiddleware(defaultMiddleware(handleAuthenticate(validate, logger, metrics, db, hmacSecret))))

	return mux
}

func AddPprofRoutes(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
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
