package services

import (
	"errors"
	"fmt"
	"sentinel-id/internal/models"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/utils"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	UserRepo    *repositories.UserRepository
	SessionRepo *repositories.SessionRepository
	AuditRepo   *repositories.AuditRepository
	QRRepo      *repositories.QRRepository
}

func NewUserService(
	userRepo *repositories.UserRepository,
	sessionRepo *repositories.SessionRepository,
	auditRepo *repositories.AuditRepository,
	qrRepo *repositories.QRRepository,
) *UserService {
	return &UserService{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		AuditRepo:   auditRepo,
		QRRepo:      qrRepo,
	}
}

func (s *UserService) Register(input models.UserRegisterDTO) (*models.User, error) {
	existingUser, _ := s.UserRepo.FindByEmail(input.Email)
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:           uuid.New(),
		FullName:     input.FullName,
		Email:        input.Email,
		CPF:          input.CPF,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		MfaEnabled:   false,
	}

	err = s.UserRepo.CreateUser(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) Login(input models.UserLoginDTO, userAgent string, clientIP string) (string, string, error) {
	user, err := s.UserRepo.FindByEmail(input.Email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	match, err := utils.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil || !match {
		s.AuditRepo.Create(models.AuditLog{
			ID:        uuid.New(),
			UserID:    user.ID,
			Action:    "LOGIN_FAILED",
			IPAddress: clientIP,
			UserAgent: userAgent,
			Details:   `{"reason": "invalid_password"}`,
			CreatedAt: time.Now(),
		})
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := utils.GenerateToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	session := models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
		DeviceInfo:   userAgent,
		IPAddress:    clientIP,
		ExpiresAt:    time.Now().Add(24 * time.Hour * 7),
		CreatedAt:    time.Now(),
	}

	err = s.SessionRepo.Create(session)
	if err != nil {
		return "", "", err
	}

	s.AuditRepo.Create(models.AuditLog{
		ID:        uuid.New(),
		UserID:    user.ID,
		Action:    "LOGIN_SUCCESS",
		IPAddress: clientIP,
		UserAgent: userAgent,
		Details:   "{}",
		CreatedAt: time.Now(),
	})

	return accessToken, refreshToken, nil
}

func (s *UserService) RefreshToken(refreshToken string) (string, error) {
	session, err := s.SessionRepo.FindByRefreshToken(refreshToken)
	if err != nil || session == nil {
		return "", errors.New("invalid refresh token")
	}

	if session.IsRevoked {
		return "", errors.New("session is revoked")
	}

	if time.Now().After(session.ExpiresAt) {
		return "", errors.New("session expired")
	}

	newAccessToken, err := utils.GenerateToken(session.UserID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *UserService) Logout(refreshToken string) error {
	return s.SessionRepo.Revoke(refreshToken)
}

func (s *UserService) RevokeAllSessions(userID uuid.UUID, ip string, userAgent string) error {
	err := s.SessionRepo.RevokeAll(userID)
	if err != nil {
		return err
	}

	s.AuditRepo.Create(models.AuditLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "KILL_SWITCH_ACTIVATED",
		IPAddress: ip,
		UserAgent: userAgent,
		Details:   `{"reason": "user_requested"}`,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *UserService) ListUserSessions(userID uuid.UUID) ([]models.Session, error) {
	return s.SessionRepo.ListByUser(userID)
}

func (s *UserService) LogoutAll(userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.RevokeAllSessions(uid, "Unknown", "Unknown")
}

func (s *UserService) ListSessions(userID string) ([]models.Session, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.SessionRepo.ListByUser(uid)
}

func (s *UserService) ListAuditLogs(userID uuid.UUID) ([]models.AuditLog, error) {
	return s.AuditRepo.ListByUser(userID.String())
}

func (s *UserService) RevokeSession(sessionID string, userID uuid.UUID, ip, userAgent string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return err
	}

	err = s.SessionRepo.RevokeByID(sid, userID)
	if err != nil {
		return err
	}

	s.AuditRepo.Create(models.AuditLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "SESSION_REVOKED",
		IPAddress: ip,
		UserAgent: userAgent,
		Details:   fmt.Sprintf(`{"revoked_session_id": "%s"}`, sessionID),
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *UserService) InitQRLogin() (*models.LoginRequest, error) {
	req := models.LoginRequest{
		ID:        uuid.New(),
		Status:    "PENDING",
		ExpiresAt: time.Now().Add(2 * time.Minute),
		CreatedAt: time.Now(),
	}

	err := s.QRRepo.Create(req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *UserService) ApproveQRLogin(requestID string, userID uuid.UUID) error {
	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return errors.New("invalid request ID")
	}

	req, err := s.QRRepo.FindByID(reqUUID)
	if err != nil || req == nil {
		return errors.New("login request not found")
	}
	if time.Now().After(req.ExpiresAt) {
		return errors.New("qr code expired")
	}
	if req.Status != "PENDING" {
		return errors.New("qr code already used or invalid")
	}

	return s.QRRepo.Approve(reqUUID, userID)
}

func (s *UserService) PollQRLogin(requestID string, userAgent, ip string) (string, string, error) {
	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return "", "", errors.New("invalid request ID")
	}

	req, err := s.QRRepo.FindByID(reqUUID)
	if err != nil {
		return "", "", errors.New("request not found")
	}

	if req.Status == "PENDING" {
		return "", "", nil
	}

	if req.Status == "USED" || time.Now().After(req.ExpiresAt) {
		return "", "", errors.New("expired or used")
	}

	if req.Status == "APPROVED" && req.UserID != nil {
		s.QRRepo.MarkAsUsed(reqUUID)

		accessToken, err := utils.GenerateToken(*req.UserID)
		if err != nil {
			return "", "", err
		}

		refreshToken, err := utils.GenerateRefreshToken(*req.UserID)
		if err != nil {
			return "", "", err
		}

		session := models.Session{
			ID:           uuid.New(),
			UserID:       *req.UserID,
			Token:        accessToken,
			RefreshToken: refreshToken,
			DeviceInfo:   userAgent + " (via QR)",
			IPAddress:    ip,
			ExpiresAt:    time.Now().Add(24 * time.Hour * 7),
			CreatedAt:    time.Now(),
		}
		s.SessionRepo.Create(session)

		s.AuditRepo.Create(models.AuditLog{
			ID:        uuid.New(),
			UserID:    *req.UserID,
			Action:    "LOGIN_QR_SUCCESS",
			IPAddress: ip,
			UserAgent: userAgent,
			Details:   "{}",
			CreatedAt: time.Now(),
		})

		return accessToken, refreshToken, nil
	}

	return "", "", errors.New("unknown state")
}

func (s *UserService) SetTrustedDevice(userID string, deviceID string) error {
	return s.UserRepo.SetTrustedDevice(userID, deviceID)
}

func (s *UserService) InitPromptLogin(email string) (*models.LoginRequest, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("utilizador não encontrado")
	}

	if user.TrustedDeviceID == nil || *user.TrustedDeviceID == "" {
		return nil, errors.New("este utilizador não possui um dispositivo de confiança configurado")
	}

	reqID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Minute)

	err = s.QRRepo.CreatePromptRequest(reqID.String(), user.ID.String(), expiresAt)
	if err != nil {
		return nil, err
	}

	req := models.LoginRequest{
		ID:        reqID,
		Status:    "PENDING",
		ExpiresAt: expiresAt,
	}
	return &req, nil
}

func (s *UserService) CheckPendingPrompt(userID string) (string, error) {
	return s.QRRepo.GetPendingPrompt(userID)
}
