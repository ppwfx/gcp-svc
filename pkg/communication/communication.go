package communication

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"log"
	"net"
	"net/http"
)

func Serve(v *validator.Validate, db *sqlx.DB, addr string) (err error) {
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
			Password: req.Password,
			FullName: req.FullName,
		})
		if err != nil {
			rsp.Error = err.Error()
			statusCode = http.StatusUnprocessableEntity

			return
		}

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
