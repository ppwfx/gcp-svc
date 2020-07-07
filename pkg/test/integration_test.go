package test

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/jmoiron/sqlx"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var args = types.IntegrationTestArgs{}
var db *sqlx.DB

func TestMain(m *testing.M) {
	flag.StringVar(&args.UserSvcAddr, "user-svc-addr", "", "")
	flag.StringVar(&args.DbConnection, "db-connection", "", "")
	flag.Parse()

	err := func() (err error) {
		err = persistence.WaitForDb(args.DbConnection)
		if err != nil {
			return
		}

		db, err = persistence.GetDb(25, 25, 5*time.Minute, args.DbConnection)
		if err != nil {
			return
		}

		err = persistence.Migrate(db)
		if err != nil {
			return
		}

		return
	}()
	if err != nil {
		log.Fatal(err)
	}

	c := m.Run()

	os.Exit(c)
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		req := types.CreateUserRequest{
			Email:    "johndoe@example.com",
			Password: "password",
			FullName: "johndoe",
		}

		var b bytes.Buffer
		err = json.NewEncoder(&b).Encode(req)
		if err != nil {
			return
		}

		resp, err := http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
		if err != nil {
			return
		}

		var rsp types.CreateUserResponse
		err = json.NewDecoder(resp.Body).Decode(&rsp)
		if err != nil {
			return
		}

		assert.Equal(t, resp.StatusCode, http.StatusOK)
		assert.Empty(t, rsp.Error)

		u, err := persistence.GetUserByEmail(db, req.Email)
		if err != nil {
			return
		}

		assert.Equal(t, req.Email, u.Email)
		assert.NotEqual(t, req.Password, u.Password)
		assert.Equal(t, req.FullName, u.FullName)

		return
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		createReq := types.CreateUserRequest{
			Email:    "johndoe1@example.com",
			Password: "password",
			FullName: "johndoe",
		}

		var b bytes.Buffer
		err = json.NewEncoder(&b).Encode(createReq)
		if err != nil {
			return
		}

		_, err = http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
		if err != nil {
			return
		}

		authReq := types.AuthenticateRequest{
			Email:    "johndoe1@example.com",
			Password: "password",
		}

		err = json.NewEncoder(&b).Encode(authReq)
		if err != nil {
			return
		}

		resp, err := http.Post("http://"+args.UserSvcAddr+types.RouteAuthenticate, types.ContentTypeJson, &b)
		if err != nil {
			return
		}

		var rsp types.AuthenticateResponse
		err = json.NewDecoder(resp.Body).Decode(&rsp)
		if err != nil {
			return
		}

		assert.Equal(t, resp.StatusCode, http.StatusOK)
		assert.Empty(t, rsp.Error)
		assert.NotEmpty(t, rsp.AccessToken)

		return
	}()
	if err != nil {
		t.Error(err)
	}
}
