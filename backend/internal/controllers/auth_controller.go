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

func NewAuthController(service *services.UserService) *AuthController {
	return &AuthController{Service: service}
}

// @Summary      Registrar novo usuário
// @Description  Cria uma nova conta de usuário com senha criptografada
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body models.UserRegisterDTO true "Dados de Cadastro"
// @Success      201  {object} map[string]interface{}
// @Failure      400  {object} map[string]string
// @Failure      409  {object} map[string]string
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
// @Description  Autentica o usuário e retorna Access Token e Refresh Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body models.UserLoginDTO true "Credenciais"
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
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
// @Description  Gera um novo Access Token usando um Refresh Token válido
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshInput true "Refresh Token"
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
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
// @Description  Revoga um Refresh Token específico
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshInput true "Token para revogar"
// @Success      200  {object} map[string]string
// @Failure      400  {object} map[string]string
// @Failure      500  {object} map[string]string
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
// @Description  Desconecta o usuário de TODOS os dispositivos
// @Tags         Users
// @Security     Bearer
// @Success      200  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Failure      500  {object} map[string]string
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
// @Description  Retorna todas as sessões ativas e identifica a atual
// @Tags         Users
// @Security     Bearer
// @Success      200  {array} SessionResponse
// @Failure      401  {object} map[string]string
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
