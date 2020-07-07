package communication

import (
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"log"
	"net"
	"net/http"
)

func Serve(v *validator.Validate, db *sqlx.DB, addr string, hmacSecret string, salt string, allowedSubjectSuffix string) (err error) {
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

	m.HandleFunc(types.RouteListUsers, func(w http.ResponseWriter, r *http.Request) {
		rsp := types.ListUsersResponse{}
		statusCode := http.StatusOK

		defer func() {
			w.WriteHeader(statusCode)

			err := json.NewEncoder(w).Encode(&rsp)
			if err != nil {
				log.Println(err)

				return
			}
		}()

		spew.Dump(r.Header)

		accessToken := business.ExtractAccessToken(r)
		if accessToken == "" {
			rsp.Error = types.ErrorUnauthorized
			statusCode = http.StatusUnauthorized

			return
		}

		is, err := business.IsAuthorized(hmacSecret, accessToken, allowedSubjectSuffix)
		if err != nil {
			log.Println(err)

			statusCode = http.StatusUnprocessableEntity

			return
		}
		if !is {
			rsp.Error = types.ErrorUnauthorized
			statusCode = http.StatusUnauthorized

			return
		}

		var req types.ListUsersRequest
		err = json.NewDecoder(r.Body).Decode(&req)
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

		us, err := persistence.SelectUsersOrderByIdDesc(db)
		if err != nil {
			rsp.Error = types.ErrorInternalError
			statusCode = http.StatusInternalServerError

			return
		}

		for _, u := range us {
			rsp.Users = append(rsp.Users, types.ListUser{
				Id:       u.Id,
				Email:    u.Email,
				FullName: u.FullName,
			})
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

		accessToken, err := business.GenerateAccessToken(hmacSecret, u.Email)
		if err != nil {
			log.Println(err)

			rsp.Error = types.ErrorInternalError
			statusCode = http.StatusInternalServerError

			return
		}

		rsp.AccessToken = accessToken

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
