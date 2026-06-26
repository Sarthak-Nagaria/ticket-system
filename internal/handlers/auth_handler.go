package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Sarthak-Nagaria/ticket-system/internal/config"
	"github.com/Sarthak-Nagaria/ticket-system/internal/models"
	"github.com/Sarthak-Nagaria/ticket-system/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler holds dependencies required by authentication endpoints.
type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Cfg: cfg}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request payload: "+err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	if password == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "password cannot be empty")
		return
	}

	var existing models.User
	err := h.DB.Where("email = ?", email).First(&existing).Error

	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "email already registered")
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to check existing user")
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		Email:        email,
		PasswordHash: hash,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request payload: "+err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	var user models.User

	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "invalid email or password")
			return
		}

		utils.ErrorResponse(c, http.StatusInternalServerError, "database error")
		return
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := utils.GenerateToken(
		user.ID,
		user.Email,
		h.Cfg.JWTSecret,
		h.Cfg.JWTExpiryHours,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"token":            token,
		"token_type":       "Bearer",
		"expires_in_hours": h.Cfg.JWTExpiryHours,
	})
}