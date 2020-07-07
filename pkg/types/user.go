package types

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"email"`
	FullName string `json:"full_name"`
}

type CreateUserResponse struct {
	Error string `json:"email"`
}
