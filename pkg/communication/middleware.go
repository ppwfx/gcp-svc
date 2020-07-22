package communication

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils"
	"go.uber.org/zap"
)

func composeAuthMiddleware(hmacSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := extractAccessToken(r)

		claims, err := business.GetJwtClaims(hmacSecret, t)
		if err != nil {
			utils.GetContextLogger(r.Context()).With(
				types.LogFunc, "composeAuthMiddleware",
			).Error(err)

			writeJsonResponse(w, http.StatusUnauthorized, types.ErrorResponse{
				Error: types.ErrorUnauthorized,
			})

			return
		}

		l := utils.GetContextLogger(r.Context()).With(
			types.LogSub, claims[types.ClaimSub],
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

		writeJsonResponse(w, http.StatusForbidden, types.ErrorResponse{
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
	count int
	code  int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}

func (iw *interceptingWriter) Write(p []byte) (int, error) {
	iw.count += len(p)
	return iw.ResponseWriter.Write(p)
}

func composeContextLoggerMiddleware(l *zap.SugaredLogger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		iw := &interceptingWriter{0, http.StatusOK, w}

		l := l.With(
			"req_id", uuid.New().String(),
		)

		r = r.WithContext(utils.WithContextLogger(r.Context(), l))

		begin := time.Now()

		next(iw, r)

		l.With(
			"http_req_remoteaddr", r.RemoteAddr,
			"http_req_method", r.Method,
			"http_req_url", r.URL.String(),
			"http_req_contentlength", r.ContentLength,
			"http_resp_statuscode", iw.code,
			"http_resp_statustext", http.StatusText(iw.code),
			"http_resp_size", iw.count,
			"http_resp_took", time.Since(begin).String(),
			"http_resp_sec", time.Since(begin).Seconds(),
		).Info()
	}
}