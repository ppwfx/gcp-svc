package types

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	FullName string `json:"fullname" validate:"required"`
}

type CreateUserResponse struct {
	Error string `json:"error"`
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
	Id       int    `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	FullName string `db:"fullname"`
}
