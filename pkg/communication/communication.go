package communication

import (
	"encoding/json"
	"github.com/ppwfx/user-svc/pkg/types"
	"log"
	"net"
	"net/http"
)

func Serve(addr string) (err error) {
	m := http.NewServeMux()

	m.HandleFunc(types.RouteCreateUser, func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateUserRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		rsp := types.CreateUserResponse{}

		err = json.NewEncoder(w).Encode(&rsp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			log.Println(err)

			return
		}
	})

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	err = http.Serve(l, m)
	if err != nil {
		return
	}

	return
}
