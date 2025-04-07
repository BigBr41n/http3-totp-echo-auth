package dtos

type CreateUserDTO struct {
	Username string `json:"username" validate:"required,min=4,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,regexp=^.*(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[!@#$%^&*()_+]).*$"`
	Role     string `json:"role" validate:"required,oneof=client seller investor"`
}
