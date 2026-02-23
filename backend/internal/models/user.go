package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID `json:"id"`
	FullName        string    `json:"full_name"`
	Email           string    `json:"email"`
	CPF             string    `json:"cpf"`
	PasswordHash    string    `json:"-"`
	MfaEnabled      bool      `json:"mfa_enabled"`
	TrustedDeviceID *string   `json:"trusted_device_id" db:"trusted_device_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type UserRegisterDTO struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	CPF      string `json:"cpf" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserLoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
