// +build integration

package communication

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/ppwfx/user-svc/pkg/business"
	"github.com/ppwfx/user-svc/pkg/communication/client"
	"github.com/ppwfx/user-svc/pkg/persistence"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/ppwfx/user-svc/pkg/utils/ctxutil"
	"github.com/ppwfx/user-svc/pkg/utils/dockerutil"
	"github.com/ppwfx/user-svc/pkg/utils/metricsutil"
)

var args = types.IntegrationTestArgs{}
var db *sqlx.DB
var ctx = context.Background()
var httpClient *http.Client
var userSvcAddr string
var pgUrl string
var metricSink metrics.MetricSink
var prefix = time.Now().Format("2006-01-02T15-04-05")

func TestMain(m *testing.M) {
	flag.BoolVar(&args.Remote, "remote", false, "")
	flag.StringVar(&args.UserSvcAddr, "user-svc-addr", "", "")
	flag.StringVar(&args.PostgresUrl, "postgres-url", "", "")
	flag.Parse()

	if args.Remote {
		userSvcAddr = args.UserSvcAddr
		pgUrl = args.PostgresUrl
		httpClient = &http.Client{}
	} else {
		pgUrl = "postgres://user:password@localhost:5432/user-svc?sslmode=disable"
	}

	logger := zap.NewNop().Sugar()

	ctx = ctxutil.WithContextLogger(ctx, logger)

	err := func() (err error) {
		if !args.Remote {
			err = dockerutil.RemoveDockerContainers("user-svc-communication")
			if err != nil {
				return
			}

			var o []byte
			o, err = exec.Command("docker", strings.Fields("run -d --label user-svc-communication --rm -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=user-svc -p 5432:5432 postgres")...).Output()
			if err != nil {
				err = errors.Wrapf(err, "failed to run postgres container: %s", o)

				return
			}

			db, err = persistence.OpenPostgresDB(25, 25, 5*time.Minute, pgUrl)
			if err != nil {
				err = errors.Wrapf(err, "failed to open postgres")

				return
			}

			err = persistence.ConnectToPostgresDb(ctx, db, 10*time.Second)
			if err != nil {
				err = errors.Wrapf(err, "failed to connect to postgres")

				return
			}

			err = persistence.Migrate(logger, "file://../persistence/migrations", pgUrl)
			if err != nil {
				err = errors.Wrapf(err, "failed to migrate postgres")

				return
			}

			validate := validator.New()

			metricSink, err = metricsutil.NewInMemoryMetricSink()
			if err != nil {
				err = errors.Wrap(err, "failed to get development metrics")

				return
			}

			go func() {
				mux := http.NewServeMux()
				mux = AddSvcRoutes(mux, validate, logger, metricSink, db, "hmac-secret", "@test.com", business.DefaultArgon2IdOpts)

				testServer := httptest.NewServer(mux)
				httpClient = testServer.Client()

				userSvcAddr = testServer.URL
			}()
		}

		return
	}()
	if err != nil {
		log.Fatal(err)
	}

	c := m.Run()

	if !args.Remote {
		err = dockerutil.RemoveDockerContainers("user-svc-communication")
		if err != nil {
			err = errors.Wrapf(err, "failed to remove docker containers")

			log.Fatal(err)
		}
	}

	os.Exit(c)
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name               string
		firstCreateReq     types.CreateUserRequest
		secondCreateReq    types.CreateUserRequest
		expectError        bool
		expectedStatusCode int
	}{
		{
			name: "valid user",
			secondCreateReq: types.CreateUserRequest{
				Email:    prefix + "testCreateUser0@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			expectError:        false,
			expectedStatusCode: 200,
		},
		{
			name: "invalid user without unique email",
			firstCreateReq: types.CreateUserRequest{
				Email:    prefix + "testCreateUser1@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			secondCreateReq: types.CreateUserRequest{
				Email:    prefix + "testCreateUser1@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
		{
			name: "invalid user without valid email",
			secondCreateReq: types.CreateUserRequest{
				Email:    prefix + "testCreateUser2@example",
				Password: "password",
				FullName: "johndoe",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
		{
			name: "invalid user without email",
			secondCreateReq: types.CreateUserRequest{
				Email:    "",
				Password: "password",
				FullName: "johndoe",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
		{
			name: "invalid user without password",
			secondCreateReq: types.CreateUserRequest{
				Email:    prefix + "testCreateUser4@example.com",
				Password: "",
				FullName: "johndoe",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
		{
			name: "invalid user without fullname",
			secondCreateReq: types.CreateUserRequest{
				Email:    prefix + "testCreateUser5@example.com",
				Password: "password",
				FullName: "",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := func() (err error) {
				_, _, _ = client.CreateUser(ctx, httpClient, userSvcAddr, tc.firstCreateReq)

				httpRsp, createRsp, err := client.CreateUser(ctx, httpClient, userSvcAddr, tc.secondCreateReq)
				if err != nil {
					return
				}

				assert.Equal(t, tc.expectedStatusCode, httpRsp.StatusCode)

				if tc.expectError {
					assert.NotEmpty(t, createRsp.Error)
				} else {
					assert.Empty(t, createRsp.Error)
				}

				if pgUrl != "" && !tc.expectError {
					var u types.UserModel
					u, err = persistence.GetUserByEmail(ctx, metricSink, db, tc.secondCreateReq.Email)
					if err != nil {
						return
					}

					_, err := uuid.Parse(u.ID)
					assert.NoError(t, err)
					assert.Equal(t, tc.secondCreateReq.Email, u.Email)
					assert.NotEqual(t, tc.secondCreateReq.Password, u.Password)
					assert.Equal(t, tc.secondCreateReq.FullName, u.FullName)
					diff := time.Now().Sub(u.CreatedAt)
					assert.True(t, -time.Second < diff && diff < time.Second, "expected difference to be between -1 and 1 second, actual: %v", diff)
				}

				return
			}()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name               string
		createReq          types.CreateUserRequest
		authReq            types.AuthenticateRequest
		expectError        bool
		expectedStatusCode int
	}{
		{
			name: "valid credentials",
			createReq: types.CreateUserRequest{
				Email:    prefix + "testAuthenticate0@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testAuthenticate0@example.com",
				Password: "password",
			},
			expectError:        false,
			expectedStatusCode: 200,
		},
		{
			name: "invalid credentials without password",
			createReq: types.CreateUserRequest{
				Email:    prefix + "testAuthenticate1@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testAuthenticate1@example.com",
				Password: "",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
		{
			name: "invalid credentials without same password",
			createReq: types.CreateUserRequest{
				Email:    prefix + "testAuthenticate2@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testAuthenticate2@example.com",
				Password: "password2",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
		{
			name: "invalid credentials without existing user",
			createReq: types.CreateUserRequest{
				Email:    "",
				Password: "",
				FullName: "",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testAuthenticate3@example.com",
				Password: "password2",
			},
			expectError:        true,
			expectedStatusCode: 422,
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := func() (err error) {
				_, _, _ = client.CreateUser(ctx, httpClient, userSvcAddr, tc.createReq)

				httpRsp, authRsp, err := client.Authenticate(ctx, httpClient, userSvcAddr, tc.authReq)
				if err != nil {
					return
				}

				assert.Equal(t, tc.expectedStatusCode, httpRsp.StatusCode)

				if tc.expectError {
					assert.NotEmpty(t, authRsp.Error)
					assert.Empty(t, authRsp.AccessToken)
				} else {
					assert.Empty(t, authRsp.Error)
					assert.NotEmpty(t, authRsp.AccessToken)
				}

				return
			}()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

type testListUsersTestCase struct {
	name               string
	createReq          types.CreateUserRequest
	authReq            types.AuthenticateRequest
	expectError        bool
	expectedStatusCode int
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name               string
		createReq          types.CreateUserRequest
		authReq            types.AuthenticateRequest
		expectError        bool
		expectedStatusCode int
	}{
		{
			name: "valid user",
			createReq: types.CreateUserRequest{
				Email:    prefix + "testListUsers0@test.com",
				Password: "password",
				FullName: "johndoe",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testListUsers0@test.com",
				Password: "password",
			},
			expectError:        false,
			expectedStatusCode: 200,
		},
		{
			name: "invalid user without permissions",
			createReq: types.CreateUserRequest{
				Email:    prefix + "testListUsers1@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testListUsers1@example.com",
				Password: "password",
			},
			expectError:        true,
			expectedStatusCode: 403,
		},
	}

	for _, tc := range tcs {
		testListUsers(t, tc)
	}
}

func testListUsers(t *testing.T, tc testListUsersTestCase) {
	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()

		err := func() (err error) {
			_, _, _ = client.CreateUser(ctx, httpClient, userSvcAddr, tc.createReq)

			_, authRsp, _ := client.Authenticate(ctx, httpClient, userSvcAddr, tc.authReq)

			httpRsp, listRsp, err := client.ListUsers(ctx, httpClient, userSvcAddr, authRsp.AccessToken, types.ListUsersRequest{})
			if err != nil {
				return
			}

			assert.Equal(t, tc.expectedStatusCode, httpRsp.StatusCode)

			if tc.expectError {
				assert.NotEmpty(t, listRsp.Error)
				assert.Len(t, listRsp.Users, 0)
			} else {
				assert.Empty(t, listRsp.Error)
				assert.NotEmpty(t, listRsp.Users)
			}

			return
		}()
		if err != nil {
			t.Error(err)
		}
	})
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name               string
		firstCreateReq     types.CreateUserRequest
		secondCreateReq    types.CreateUserRequest
		authReq            types.AuthenticateRequest
		deleteReq          types.DeleteUserRequest
		expectError        bool
		expectedStatusCode int
	}{
		{
			name: "valid user",
			firstCreateReq: types.CreateUserRequest{
				Email:    prefix + "testDeleteUser0@test.com",
				Password: "password",
				FullName: "johndoe",
			},
			secondCreateReq: types.CreateUserRequest{
				Email:    prefix + "testDeleteUser0@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testDeleteUser0@test.com",
				Password: "password",
			},
			deleteReq: types.DeleteUserRequest{
				Email: prefix + "testDeleteUser0@example.com",
			},
			expectError:        false,
			expectedStatusCode: 200,
		},
		{
			name: "invalid user without permissions",
			firstCreateReq: types.CreateUserRequest{
				Email:    prefix + "testDeleteUser1@test.com",
				Password: "password",
				FullName: "johndoe",
			},
			secondCreateReq: types.CreateUserRequest{
				Email:    prefix + "testDeleteUser1@example.com",
				Password: "password",
				FullName: "johndoe",
			},
			authReq: types.AuthenticateRequest{
				Email:    prefix + "testDeleteUser1@test.com",
				Password: "password1",
			},
			deleteReq: types.DeleteUserRequest{
				Email: prefix + "testDeleteUser1@example.com",
			},
			expectError:        true,
			expectedStatusCode: 401,
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := func() (err error) {
				_, _, _ = client.CreateUser(ctx, httpClient, userSvcAddr, tc.firstCreateReq)

				_, _, _ = client.CreateUser(ctx, httpClient, userSvcAddr, tc.secondCreateReq)

				_, authRsp, _ := client.Authenticate(ctx, httpClient, userSvcAddr, tc.authReq)

				httpRsp, deleteRsp, err := client.DeleteUser(ctx, httpClient, userSvcAddr, authRsp.AccessToken, tc.deleteReq)
				if err != nil {
					return
				}

				assert.Equal(t, tc.expectedStatusCode, httpRsp.StatusCode)

				if pgUrl != "" {
					var u types.UserModel
					if tc.expectError {
						u, err = persistence.GetUserByEmail(ctx, metricSink, db, tc.deleteReq.Email)
						if err != nil {
							return
						}

						assert.Equal(t, tc.deleteReq.Email, u.Email)
					} else {
						var u types.UserModel
						u, err = persistence.GetUserByEmail(ctx, metricSink, db, tc.deleteReq.Email)

						assert.Error(t, err)
						err = nil

						assert.Equal(t, "", u.Email)
					}
				}

				if tc.expectError {
					assert.NotEmpty(t, deleteRsp.Error)
				} else {
					assert.Empty(t, deleteRsp.Error)
				}

				return
			}()
			if err != nil {
				t.Error(err)
			}
		})
	}
}
