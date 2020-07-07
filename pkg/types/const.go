package types

const (
	RouteCreateUser         = "/api/v0/createUser"
	RouteListUsers          = "/api/v0/listUsers"
	RouteAuthenticate       = "/api/v0/authenticate"
	ContentTypeJson         = "application/json"
	ErrorInvalidCredentials = "invalid credentials"
	ErrorInternalError      = "internal error"
	ErrorUnauthorized       = "unauthorized"
	HeaderAuthorization     = "Authorization"
	HeaderContentType       = "Content-Type"
	PrefixBearer            = "Bearer "
	ClaimSub                = "sub"
)
