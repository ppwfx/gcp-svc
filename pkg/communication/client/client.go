package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/types"
)

func CreateUser(ctx context.Context, c *http.Client, addr string, req types.CreateUserRequest) (httpRsp *http.Response, rsp types.CreateUserResponse, err error) {
	httpRsp, err = do(ctx, c, addr, types.RouteCreateUser, "", req, &rsp)
	if err != nil {
		return
	}

	return
}

func Authenticate(ctx context.Context, c *http.Client, addr string, req types.AuthenticateRequest) (httpRsp *http.Response, rsp types.AuthenticateResponse, err error) {
	httpRsp, err = do(ctx, c, addr, types.RouteAuthenticate, "", req, &rsp)
	if err != nil {
		return
	}

	return
}

func ListUsers(ctx context.Context, c *http.Client, addr string, token string, req types.ListUsersRequest) (httpRsp *http.Response, rsp types.ListUsersResponse, err error) {
	httpRsp, err = do(ctx, c, addr, types.RouteListUsers, token, req, &rsp)
	if err != nil {
		return
	}

	return
}

func DeleteUser(ctx context.Context, c *http.Client, addr string, token string, req types.DeleteUserRequest) (httpRsp *http.Response, rsp types.DeleteUserResponse, err error) {
	httpRsp, err = do(ctx, c, addr, types.RouteDeleteUser, token, req, &rsp)
	if err != nil {
		return
	}

	return
}

func do(ctx context.Context, c *http.Client, addr string, path string, token string, req interface{}, rsp interface{}) (httpRsp *http.Response, err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(req)
	if err != nil {
		return
	}

	var r *http.Request
	r, err = http.NewRequest(http.MethodPost, addr+path, &buf)
	if err != nil {
		return
	}
	r.Header.Set(types.HeaderContentType, types.ContentTypeJson)
	if token != "" {
		r.Header.Set(types.HeaderAuthorization, types.PrefixBearer+token)
	}

	httpRsp, err = c.Do(r)
	if err != nil {
		return
	}

	b, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return
	}
	defer httpRsp.Body.Close()

	err = json.Unmarshal(b, &rsp)
	if err != nil {
		err = errors.Wrapf(err, "failed to unmarshal json: %s", b)

		return
	}

	return
}
