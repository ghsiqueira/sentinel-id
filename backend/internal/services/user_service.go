package services

import (
	"errors"
	"sentinel-id/internal/models"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/utils"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	Repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) Register(input models.UserRegisterDTO) (*models.User, error) {
	existingUser, err := s.Repo.GetByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	existingUser, err = s.Repo.GetByCPF(input.CPF)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("cpf already registered")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:           uuid.New(),
		FullName:     input.FullName,
		Email:        input.Email,
		CPF:          input.CPF,
		PasswordHash: hashedPassword,
		MfaEnabled:   false,
		CreatedAt:    time.Now(),
	}

	err = s.Repo.CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}
