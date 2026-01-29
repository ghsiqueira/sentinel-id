package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	CPF          string    `json:"cpf"`
	PasswordHash string    `json:"-"`
	MfaEnabled   bool      `json:"mfa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
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
