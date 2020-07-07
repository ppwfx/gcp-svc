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
		createReq := types.CreateUserRequest{
			Email:    "johndoe@example.com",
			Password: "password",
			FullName: "johndoe",
		}

		var b bytes.Buffer
		err = json.NewEncoder(&b).Encode(createReq)
		if err != nil {
			return
		}

		resp, err := http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
		if err != nil {
			return
		}

		var createRsp types.CreateUserResponse
		err = json.NewDecoder(resp.Body).Decode(&createRsp)
		if err != nil {
			return
		}

		assert.Equal(t, resp.StatusCode, http.StatusOK)
		assert.Empty(t, createRsp.Error)

		u, err := persistence.GetUserByEmail(db, createReq.Email)
		if err != nil {
			return
		}

		assert.Equal(t, createReq.Email, u.Email)
		assert.NotEqual(t, createReq.Password, u.Password)
		assert.Equal(t, createReq.FullName, u.FullName)

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

		var authRsp types.AuthenticateResponse
		err = json.NewDecoder(resp.Body).Decode(&authRsp)
		if err != nil {
			return
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Empty(t, authRsp.Error)
		assert.NotEmpty(t, authRsp.AccessToken)

		return
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		createReq := types.CreateUserRequest{
			Email:    "johndoe2@test.com",
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
			Email:    createReq.Email,
			Password: createReq.Password,
		}

		err = json.NewEncoder(&b).Encode(authReq)
		if err != nil {
			return
		}

		resp, err := http.Post("http://"+args.UserSvcAddr+types.RouteAuthenticate, types.ContentTypeJson, &b)
		if err != nil {
			return
		}

		var authRsp types.AuthenticateResponse
		err = json.NewDecoder(resp.Body).Decode(&authRsp)
		if err != nil {
			return
		}

		listReq := types.ListUsersRequest{}

		err = json.NewEncoder(&b).Encode(listReq)
		if err != nil {
			return
		}

		requ, err := http.NewRequest(http.MethodPost, "http://"+args.UserSvcAddr+types.RouteListUsers, &b)
		if err != nil {
			return
		}
		requ.Header.Set(types.HeaderContentType, types.ContentTypeJson)
		requ.Header.Set(types.HeaderAuthorization, types.PrefixBearer+authRsp.AccessToken)

		resp, err = http.DefaultClient.Do(requ)
		if err != nil {
			return
		}

		var listRsp types.ListUsersResponse
		err = json.NewDecoder(resp.Body).Decode(&listRsp)
		if err != nil {
			return
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Empty(t, listRsp.Error)
		assert.NotEmpty(t, listRsp.Users)

		return
	}()
	if err != nil {
		t.Error(err)
	}
}
