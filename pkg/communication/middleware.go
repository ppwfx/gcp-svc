package communication

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils"
	"go.uber.org/zap"
)

func composeAuthMiddleware(hmacSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := extractAccessToken(r)

		l := utils.GetContextLogger(r.Context())

		claims, err := business.GetJwtClaims(hmacSecret, t)
		if err != nil {
			err = errors.Wrapf(err, "failed to authenticate user: failed to get jwt claims from jwt token: %s", t)

			utils.GetContextLogger(r.Context()).Warn(err)

			writeJsonResponse(l, w, http.StatusUnauthorized, types.ErrorResponse{
				Error: types.ErrorUnauthorized,
			})

			return
		}

		l = l.With(
			types.LogUser, claims[types.ClaimSub],
			types.LogRole, claims[types.ClaimRole],
		)

		r = r.WithContext(utils.WithContextLogger(r.Context(), l))

		r = r.WithContext(context.WithValue(r.Context(), types.ContextKeyClaims, claims))

		next(w, r)
	}
}

func authorizationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var scopes []string

		c, ok := r.Context().Value(types.ContextKeyClaims).(map[string]interface{})
		if ok {
			r, ok := c[types.ClaimRole].(string)
			if !ok {
				panic("abc")
			}

			switch r {
			case types.RoleAdmin:
				scopes = types.RoleAdminScopes
			case types.RoleUser:
				scopes = types.RoleUserScopes
			}
		} else {
			scopes = types.RoleGuestScopes
		}

		for _, s := range scopes {
			if s == r.URL.Path {
				next(w, r)

				return
			}
		}

		// TODO
		//log.Println(err)

		l := utils.GetContextLogger(r.Context())

		writeJsonResponse(l, w, http.StatusForbidden, types.ErrorResponse{
			Error: types.ErrorUnauthorized,
		})
	}
}

func sensitiveMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Pragma", "no-cache")

		next(w, r)
	}
}

func secureMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		w.Header().Add("X-Content-Type-Options", "nosniff")

		next(w, r)
	}
}

func composeMaxBodyBytesMiddleware(n int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, n)

		next(w, r)
	}
}

type interceptingWriter struct {
	count int64
	code  int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}

func (iw *interceptingWriter) Write(p []byte) (int, error) {
	iw.count += int64(len(p))
	return iw.ResponseWriter.Write(p)
}

func composeContextLoggerMiddleware(l *zap.SugaredLogger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		iw := &interceptingWriter{0, http.StatusOK, w}

		l := l.With(
			types.LogId, uuid.New().String(),
		)

		r = r.WithContext(utils.WithContextLogger(r.Context(), l))

		begin := time.Now()

		next(iw, r)

		l.With(zap.Object(types.LogHttpRequest, &utils.LogHttpRequest{
			Method:             r.Method,
			URL:                r.URL.String(),
			UserAgent:          r.UserAgent(),
			Referrer:           r.Referer(),
			RemoteIP:           r.RemoteAddr,
			RequestSize:        r.ContentLength,
			ResponseSize:       iw.count,
			ResponseStatusCode: iw.code,
			Latency:            fmt.Sprintf("%.6fs", time.Since(begin).Seconds()),
		}))

		if iw.code >= 500 {
			l.Error()

			return
		}

		if iw.code >= 400 {
			l.Warn()

			return
		}

		l.Info()
	}
}
