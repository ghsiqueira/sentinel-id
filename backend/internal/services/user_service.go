package services

import (
	"errors"
	"sentinel-id/internal/models"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo        *repositories.UserRepository
	SessionRepo *repositories.SessionRepository
}

func NewUserService(repo *repositories.UserRepository, sessionRepo *repositories.SessionRepository) *UserService {
	return &UserService{Repo: repo, SessionRepo: sessionRepo}
}

func (s *UserService) Register(input models.UserRegisterDTO) (*models.User, error) {
	if !utils.IsCPFValid(input.CPF) {
		return nil, errors.New("o CPF informado é inválido")
	}

	if !utils.IsPasswordStrong(input.Password) {
		return nil, errors.New("a senha é muito fraca. Requisitos: min 8 chars, maiúscula, minúscula, número e símbolo")
	}

	existingUser, _ := s.Repo.FindByEmail(input.Email)
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FullName:     input.FullName,
		Email:        input.Email,
		CPF:          input.CPF,
		PasswordHash: string(hashedPassword),
	}

	err = s.Repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(input models.UserLoginDTO, userAgent string, ip string) (string, string, error) {
	user, err := s.Repo.GetByEmail(input.Email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("invalid credentials")
	}

	match, err := utils.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil {
		return "", "", err
	}
	if !match {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		DeviceName:   userAgent,
		IPAddress:    ip,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7),
	}

	err = s.SessionRepo.CreateSession(session)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) RefreshToken(refreshTokenString string) (string, error) {
	claims, err := utils.ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	if claims["type"] != "refresh" {
		return "", errors.New("invalid token type")
	}

	session, err := s.SessionRepo.GetSessionByToken(refreshTokenString)
	if err != nil {
		return "", errors.New("session not found")
	}

	if session.IsRevoked {
		return "", errors.New("session revoked")
	}

	if time.Now().After(session.ExpiresAt) {
		return "", errors.New("session expired")
	}

	newAccessToken, err := utils.GenerateAccessToken(session.UserID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *UserService) Logout(refreshTokenString string) error {
	return s.SessionRepo.RevokeSession(refreshTokenString)
}

func (s *UserService) LogoutAll(userID string) error {
	return s.SessionRepo.RevokeAllUserSessions(userID)
}

func (s *UserService) ListSessions(userID string) ([]models.Session, error) {
	return s.SessionRepo.GetSessionsByUserID(userID)
}
