package controllers

import (
	"net/http"
	"sentinel-id/internal/models"
	"sentinel-id/internal/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	Service *services.UserService
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type SessionResponse struct {
	ID         uuid.UUID `json:"id"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
	IsCurrent  bool      `json:"is_current"`
}

type ApproveQRInput struct {
	RequestID string `json:"request_id" binding:"required"`
}

type SetTrustedDeviceInput struct {
	DeviceID string `json:"device_id" binding:"required"`
}

type InitPromptInput struct {
	Email string `json:"email" binding:"required,email"`
}

func NewAuthController(service *services.UserService) *AuthController {
	return &AuthController{Service: service}
}

// @Summary      Registrar novo usuário
// @Router       /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var input models.UserRegisterDTO

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.Service.Register(input)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

// @Summary      Realizar Login
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var input models.UserLoginDTO

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userAgent := ctx.GetHeader("User-Agent")
	clientIP := ctx.ClientIP()

	accessToken, refreshToken, err := c.Service.Login(input, userAgent, clientIP)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}

// @Summary      Renovar Token (Refresh)
// @Router       /auth/refresh [post]
func (c *AuthController) Refresh(ctx *gin.Context) {
	var input RefreshInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, err := c.Service.RefreshToken(input.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"expires_in":   900,
	})
}

// @Summary      Logout
// @Router       /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	var input RefreshInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.Service.Logout(input.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// @Summary      Kill Switch (Revogar Tudo)
// @Router       /users/revoke-all [post]
func (c *AuthController) LogoutAll(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uuid.UUID)

	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	err := c.Service.RevokeAllSessions(userID, clientIP, userAgent)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke sessions"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "All sessions revoked (Kill Switch activated)"})
}

// @Summary      Listar Sessões
// @Router       /users/sessions [get]
func (c *AuthController) ListSessions(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uuid.UUID)
	authHeader := ctx.GetHeader("Authorization")
	currentToken := strings.TrimPrefix(authHeader, "Bearer ")

	sessions, err := c.Service.ListUserSessions(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
		return
	}

	var response []SessionResponse
	for _, s := range sessions {
		response = append(response, SessionResponse{
			ID:         s.ID,
			DeviceInfo: s.DeviceInfo,
			IPAddress:  s.IPAddress,
			CreatedAt:  s.CreatedAt,
			IsCurrent:  s.Token == currentToken,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Get Audit Logs
// @Router       /users/audit-logs [get]
func (c *AuthController) GetAuditLogs(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uuid.UUID)

	logs, err := c.Service.ListAuditLogs(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

// @Summary      Revoke specific session
// @Router       /users/sessions/:id [delete]
func (c *AuthController) RevokeSession(ctx *gin.Context) {
	sessionID := ctx.Param("id")
	userID := ctx.MustGet("userID").(uuid.UUID)
	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	err := c.Service.RevokeSession(sessionID, userID, clientIP, userAgent)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Session revoked"})
}

// NOVOS ENDPOINTS QR CODE

// @Summary      Iniciar Login via QR Code
// @Router       /auth/qr/init [post]
func (c *AuthController) InitQR(ctx *gin.Context) {
	req, err := c.Service.InitQRLogin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init QR"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"request_id": req.ID,
		"expires_at": req.ExpiresAt,
	})
}

// @Summary      Aprovar Login (Mobile)
// @Router       /users/qr/approve [post]
func (c *AuthController) ApproveQR(ctx *gin.Context) {
	var input ApproveQRInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.MustGet("userID").(uuid.UUID)

	err := c.Service.ApproveQRLogin(input.RequestID, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Login approved successfully"})
}

// @Summary      Verificar Status QR (Polling)
// @Router       /auth/qr/poll/:code [get]
func (c *AuthController) PollQR(ctx *gin.Context) {
	code := ctx.Param("code")
	userAgent := ctx.GetHeader("User-Agent")
	clientIP := ctx.ClientIP()

	accessToken, refreshToken, err := c.Service.PollQRLogin(code, userAgent, clientIP)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if accessToken == "" {
		ctx.JSON(http.StatusAccepted, gin.H{"status": "pending"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// @Summary      Definir o celular atual como "Dispositivo de Confiança"
// @Router       /users/trusted-device [post]
func (c *AuthController) SetTrustedDevice(ctx *gin.Context) {
	var input SetTrustedDeviceInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.MustGet("userID").(uuid.UUID)

	err := c.Service.SetTrustedDevice(userID.String(), input.DeviceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set trusted device"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Trusted device set successfully"})
}

// @Summary      Iniciar Login via Prompt (Apenas com Email)
// @Router       /auth/prompt/init [post]
func (c *AuthController) InitPrompt(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email inválido ou ausente"})
		return
	}

	ipAddress := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	reqData, err := c.Service.InitPromptLogin(req.Email, ipAddress, userAgent)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Pedido de login enviado",
		"request_id": reqData.ID,
		"expires_at": reqData.ExpiresAt,
	})
}

func (c *AuthController) CheckPendingPrompt(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uuid.UUID)

	reqID, devInfo, ipAddress, err := c.Service.CheckPendingPrompt(userID.String())
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Nenhum pedido pendente"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"request_id":  reqID,
		"device_info": devInfo,
		"ip_address":  ipAddress,
	})
}
