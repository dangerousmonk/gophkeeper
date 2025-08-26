package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Login        string    `json:"login"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  time.Time `json:"last_login_at"`
	Active       bool      `json:"active"`
}

type RegisterUserRequest struct {
	Login          string `json:"login" validate:"required,min=3,max=150"`
	Password       string `json:"password" validate:"required,min=5"`
	HashedPassword string `json:"-"`
}

type RegisterUserResponse struct {
	Login  string `json:"login"`
	Token  string `json:"token"`
	ID     int    `json:"id"`
	Sucess bool   `json:"success"`
}

type LoginUserRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=150"`
	Password string `json:"password" validate:"required,min=5"`
}

type LoginUserResponse struct {
	Token   string `json:"token"`
	Success bool   `json:"success"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" validate:"required,min=5"`
}

type ChangePasswordResponse struct {
	Success bool `json:"success"`
}
