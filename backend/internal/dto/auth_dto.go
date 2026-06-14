package dto

type RegisterRequest struct {
	Name            string `json:"name" validate:"required,min=3"`
	Email           string `json:"email" validate:"required,email"`
	Username        string `json:"username" validate:"required,min=3"`
	Password        string `json:"password" validate:"required,min=8,max=72"`
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
}

type RegisterResponse struct {
	ID int64 `json:"id"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	CurrentPassword    string `json:"current_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=8,max=72"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required,eqfield=NewPassword"`
}

type ChangePasswordResponse struct {
	Token string `json:"token"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}
