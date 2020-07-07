package test

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/ppwfx/user-svc/pkg/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

var integrationTestArgs = types.IntegrationTestArgs{}

func TestMain(m *testing.M) {
	flag.StringVar(&integrationTestArgs.UserSvcAddr, "user-svc-addr", "", "")

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

		resp, err := http.Post(integrationTestArgs.UserSvcAddr+types.RouteCreateUser, types.ContentTypeJson, b)
		if err != nil {
			return
		}

		var rsp types.CreateUserResponse
		err = json.NewDecoder(resp.Body).Decode(rsp)
		if err != nil {
			return
		}

		assert.Empty(t, rsp.Error)

		return
	}()
	if err != nil {
		t.Error(err)
	}
}
