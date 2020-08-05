package types

const (
	RouteCreateUser               = "/api/v0/createUser"
	RouteDeleteUser               = "/api/v0/deleteUser"
	RouteListUsers                = "/api/v0/listUsers"
	RouteAuthenticate             = "/api/v0/authenticate"
	ContentTypeJson               = "application/json"
	ErrorInvalidCredentials       = "invalid credentials"
	ErrorUserDoesNotExist         = "user does not exist"
	ErrorCanNotDeleteInternalUser = "can not delete internal user"
	ErrorInternalError            = "internal error"
	ErrorUnauthorized             = "unauthorized"
	HeaderAuthorization           = "Authorization"
	HeaderContentType             = "Content-Type"
	PrefixBearer                  = "Bearer "
	ClaimExp                      = "exp"
	ClaimIat                      = "iat"
	ClaimUserGroup                = "user_group"
	ClaimSub                      = "sub"
	UserGroupUser                 = "user"
	UserGroupAdmin                = "admin"
	ContextKeyClaims              = "claims"
	LogHttpRequest                = "context.httpRequest"
	LogUser                       = "context.user"
	LogId                         = "id"
	LogRole                       = "role"
	LogLatency                    = "latency"
)

var (
	RoleGuestScopes = []string{RouteCreateUser, RouteAuthenticate}
	RoleUserScopes  = []string{}
	RoleAdminScopes = []string{RouteListUsers, RouteDeleteUser}
)
