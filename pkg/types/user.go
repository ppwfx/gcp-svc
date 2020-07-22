package types

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	FullName string `json:"fullname" validate:"required"`
}

type CreateUserResponse struct {
	Error string `json:"error"`
}

type DeleteUserRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type DeleteUserResponse struct {
	Error string `json:"error"`
}

type ListUsersRequest struct {
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListUsersResponse struct {
	Error string     `json:"error"`
	Users []ListUser `json:"users"`
}

type ListUser struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
}

type AuthenticateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthenticateResponse struct {
	Error       string `json:"error"`
	AccessToken string `json:"access_token"`
}

type UserModel struct {
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	FullName string `db:"fullname"`
	Role     string `db:"role"`
}
