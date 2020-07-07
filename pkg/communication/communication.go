package communication

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"log"
	"net"
	"net/http"
	"time"
)

func Serve(v *validator.Validate, db *sqlx.DB, addr string, hmacSecret string, salt string) (err error) {
	m := http.NewServeMux()

	m.HandleFunc(types.RouteCreateUser, func(w http.ResponseWriter, r *http.Request) {
		rsp := types.CreateUserResponse{}
		statusCode := http.StatusOK

		defer func() {
			w.WriteHeader(statusCode)

			err := json.NewEncoder(w).Encode(&rsp)
			if err != nil {
				log.Println(err)

				return
			}
		}()

		var req types.CreateUserRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			statusCode = http.StatusBadRequest

			return
		}

		err = v.Struct(&req)
		if err != nil {
			rsp.Error = err.Error()
			statusCode = http.StatusUnprocessableEntity

			return
		}

		err = persistence.InsertUser(db, types.UserModel{
			Email:    req.Email,
			Password: business.HashPassword(salt, req.Password),
			FullName: req.FullName,
		})
		if err != nil {
			rsp.Error = err.Error()
			statusCode = http.StatusUnprocessableEntity

			return
		}

		return
	})

	m.HandleFunc(types.RouteAuthenticate, func(w http.ResponseWriter, r *http.Request) {
		rsp := types.AuthenticateResponse{}
		statusCode := http.StatusOK

		defer func() {
			w.WriteHeader(statusCode)

			err := json.NewEncoder(w).Encode(&rsp)
			if err != nil {
				log.Println(err)

				return
			}
		}()

		var req types.AuthenticateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			statusCode = http.StatusBadRequest

			return
		}

		err = v.Struct(&req)
		if err != nil {
			rsp.Error = err.Error()
			statusCode = http.StatusUnprocessableEntity

			return
		}

		u, err := persistence.GetUserByEmail(db, req.Email)
		if err != nil {
			rsp.Error = types.ErrorInvalidCredentials
			statusCode = http.StatusUnprocessableEntity

			return
		}

		if business.HashPassword(salt, req.Password) != u.Password {
			rsp.Error = types.ErrorInvalidCredentials
			statusCode = http.StatusUnprocessableEntity

			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Subject:   u.Email,
		})

		tokenString, err := token.SignedString([]byte(hmacSecret))
		if err != nil {
			log.Println(err)

			rsp.Error = types.ErrorInternalError
			statusCode = http.StatusInternalServerError

			return
		}

		rsp.AccessToken = tokenString

		return
	})

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	log.Printf("listening on %v\n", addr)

	err = http.Serve(l, m)
	if err != nil {
		return
	}

	return
}
