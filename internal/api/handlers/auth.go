package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/messenger/backend/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,e164"` // e.g., +12125552368
	Password string `json:"password" binding:"required,min=8,max=72"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// TODO: Add proper error response formatting as per TZ
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := services.RegisterParams{
		Username: req.Username,
		Password: req.Password,
	}
	if req.Email != "" {
		params.Email = sql.NullString{String: req.Email, Valid: true}
	}
	if req.Phone != "" {
		params.Phone = sql.NullString{String: req.Phone, Valid: true}
	}

	user, err := h.authService.Register(c.Request.Context(), params)
	if err != nil {
		// TODO: Add proper error response formatting as per TZ
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

type loginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := services.LoginParams{
		Identifier: req.Identifier,
		Password:   req.Password,
	}

	resp, err := h.authService.Login(c.Request.Context(), params)
	if err != nil {
		// TODO: Differentiate between 404 Not Found and 401 Unauthorized
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.User.ID,
	})
}
