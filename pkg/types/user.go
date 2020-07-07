package types

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	FullName string `json:"fullname" validate:"required"`
}

type CreateUserResponse struct {
	Error string `json:"email"`
}

type UserModel struct {
	Id       int    `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	FullName string `db:"fullname"`
}
