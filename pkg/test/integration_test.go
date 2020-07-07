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
		tcs := []struct {
			createReq          types.CreateUserRequest
			expectError        bool
			expectedStatusCode int
		}{
			{
				createReq: types.CreateUserRequest{
					Email:    "testCreateUser@example.com",
					Password: "password",
					FullName: "johndoe",
				},
				expectError:        false,
				expectedStatusCode: 200,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "testCreateUser@example.com",
					Password: "password",
					FullName: "johndoe",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "johndoe",
					Password: "password",
					FullName: "johndoe",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "",
					Password: "password",
					FullName: "johndoe",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "testCreateUser@example.com",
					Password: "",
					FullName: "johndoe",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "testCreateUser@example.com",
					Password: "password",
					FullName: "",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
		}

		for _, tc := range tcs {
			createReq := tc.createReq

			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(createReq)
			if err != nil {
				return
			}

			var resp *http.Response
			resp, err = http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
			if err != nil {
				return
			}

			var createRsp types.CreateUserResponse
			err = json.NewDecoder(resp.Body).Decode(&createRsp)
			if err != nil {
				return
			}

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			if tc.expectError {
				assert.NotEmpty(t, createRsp.Error)
			} else {
				assert.Empty(t, createRsp.Error)
			}

			if !tc.expectError {
				var u types.UserModel
				u, err = persistence.GetUserByEmail(db, createReq.Email)
				if err != nil {
					return
				}

				assert.Equal(t, createReq.Email, u.Email)
				assert.NotEqual(t, createReq.Password, u.Password)
				assert.Equal(t, createReq.FullName, u.FullName)
			}
		}

		return
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		tcs := []struct {
			createReq          types.CreateUserRequest
			authReq            types.AuthenticateRequest
			expectError        bool
			expectedStatusCode int
		}{
			{
				createReq: types.CreateUserRequest{
					Email:    "testAuthenticate0@example.com",
					Password: "password",
					FullName: "johndoe",
				},
				authReq: types.AuthenticateRequest{
					Email:    "testAuthenticate0@example.com",
					Password: "password",
				},
				expectError:        false,
				expectedStatusCode: 200,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "testAuthenticate1@example.com",
					Password: "password",
					FullName: "johndoe",
				},
				authReq: types.AuthenticateRequest{
					Email:    "testAuthenticate1@example.com",
					Password: "",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "testAuthenticate2@example.com",
					Password: "password",
					FullName: "johndoe",
				},
				authReq: types.AuthenticateRequest{
					Email:    "testAuthenticate2@example.com",
					Password: "password2",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "",
					Password: "",
					FullName: "",
				},
				authReq: types.AuthenticateRequest{
					Email:    "testAuthenticate3@example.com",
					Password: "password2",
				},
				expectError:        true,
				expectedStatusCode: 422,
			},
		}

		for _, tc := range tcs {
			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(tc.createReq)
			if err != nil {
				return
			}

			_, err = http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
			if err != nil {
				return
			}

			err = json.NewEncoder(&b).Encode(tc.authReq)
			if err != nil {
				return
			}

			var resp *http.Response
			resp, err = http.Post("http://"+args.UserSvcAddr+types.RouteAuthenticate, types.ContentTypeJson, &b)
			if err != nil {
				return
			}

			var authRsp types.AuthenticateResponse
			err = json.NewDecoder(resp.Body).Decode(&authRsp)
			if err != nil {
				return
			}

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			if tc.expectError {
				assert.NotEmpty(t, authRsp.Error)
				assert.Empty(t, authRsp.AccessToken)
			} else {
				assert.Empty(t, authRsp.Error)
				assert.NotEmpty(t, authRsp.AccessToken)
			}
		}

		return
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		tcs := []struct {
			createReq          types.CreateUserRequest
			authReq            types.AuthenticateRequest
			expectError        bool
			expectedStatusCode int
		}{
			{
				createReq: types.CreateUserRequest{
					Email:    "testListUsers0@test.com",
					Password: "password",
					FullName: "johndoe",
				},
				authReq: types.AuthenticateRequest{
					Email:    "testListUsers0@test.com",
					Password: "password",
				},
				expectError:        false,
				expectedStatusCode: 200,
			},
			{
				createReq: types.CreateUserRequest{
					Email:    "testListUsers1@example.com",
					Password: "password",
					FullName: "johndoe",
				},
				authReq: types.AuthenticateRequest{
					Email:    "testListUsers1@example.com",
					Password: "password",
				},
				expectError:        true,
				expectedStatusCode: 401,
			},
		}

		for _, tc := range tcs {
			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(tc.createReq)
			if err != nil {
				return
			}

			_, err = http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
			if err != nil {
				return
			}

			err = json.NewEncoder(&b).Encode(tc.authReq)
			if err != nil {
				return
			}

			var resp *http.Response
			resp, err = http.Post("http://"+args.UserSvcAddr+types.RouteAuthenticate, types.ContentTypeJson, &b)
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

			var req *http.Request
			req, err = http.NewRequest(http.MethodPost, "http://"+args.UserSvcAddr+types.RouteListUsers, &b)
			if err != nil {
				return
			}
			req.Header.Set(types.HeaderContentType, types.ContentTypeJson)
			req.Header.Set(types.HeaderAuthorization, types.PrefixBearer+authRsp.AccessToken)

			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return
			}

			var listRsp types.ListUsersResponse
			err = json.NewDecoder(resp.Body).Decode(&listRsp)
			if err != nil {
				return
			}

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			if tc.expectError {
				assert.NotEmpty(t, listRsp.Error)
				assert.Len(t, listRsp.Users, 0)
			} else {
				assert.Empty(t, listRsp.Error)
				assert.NotEmpty(t, listRsp.Users)
			}
		}

		return
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		tcs := []struct {
			firstCreateReq     types.CreateUserRequest
			secondCreateReq    types.CreateUserRequest
			authReq            types.AuthenticateRequest
			deleteReq          types.DeleteUserRequest
			expectError        bool
			expectedStatusCode int
		}{
			{
				firstCreateReq: types.CreateUserRequest{
					Email:    "testDeleteUser0@test.com",
					Password: "password",
					FullName: "johndoe",
				},
				secondCreateReq: types.CreateUserRequest{
					Email:    "testDeleteUser0@example.com",
					Password: "password",
					FullName: "johndoe",
				},
				authReq: types.AuthenticateRequest{
					Email:    "testDeleteUser0@test.com",
					Password: "password",
				},
				deleteReq: types.DeleteUserRequest{
					Email: "testDeleteUser0@example.com",
				},
				expectError:        false,
				expectedStatusCode: 200,
			},
		}

		for _, tc := range tcs {
			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(tc.firstCreateReq)
			if err != nil {
				return
			}

			_, err = http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
			if err != nil {
				return
			}

			err = json.NewEncoder(&b).Encode(tc.secondCreateReq)
			if err != nil {
				return
			}

			_, err = http.Post("http://"+args.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, &b)
			if err != nil {
				return
			}

			err = json.NewEncoder(&b).Encode(tc.authReq)
			if err != nil {
				return
			}

			var resp *http.Response
			resp, err = http.Post("http://"+args.UserSvcAddr+types.RouteAuthenticate, types.ContentTypeJson, &b)
			if err != nil {
				return
			}

			var authRsp types.AuthenticateResponse
			err = json.NewDecoder(resp.Body).Decode(&authRsp)
			if err != nil {
				return
			}

			err = json.NewEncoder(&b).Encode(tc.deleteReq)
			if err != nil {
				return
			}

			var req *http.Request
			req, err = http.NewRequest(http.MethodPost, "http://"+args.UserSvcAddr+types.RouteDeleteUser, &b)
			if err != nil {
				return
			}
			req.Header.Set(types.HeaderContentType, types.ContentTypeJson)
			req.Header.Set(types.HeaderAuthorization, types.PrefixBearer+authRsp.AccessToken)

			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return
			}

			var deleteRsp types.DeleteUserResponse
			err = json.NewDecoder(resp.Body).Decode(&deleteRsp)
			if err != nil {
				return
			}

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			var u types.UserModel
			if tc.expectError {
				assert.NotEmpty(t, deleteRsp.Error)

				u, err = persistence.GetUserByEmail(db, tc.deleteReq.Email)
				if err != nil {
					return
				}

				assert.Equal(t, tc.deleteReq.Email, u.Email)
			} else {
				assert.Empty(t, deleteRsp.Error)

				u, err = persistence.GetUserByEmail(db, tc.deleteReq.Email)

				assert.Error(t, err)
				err = nil

				assert.Equal(t, "", u.Email)
			}
		}

		return
	}()
	if err != nil {
		t.Error(err)
	}
}
