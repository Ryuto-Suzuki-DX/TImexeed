package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"timexeed/backend/internal/auth"
	"timexeed/backend/internal/models"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResponse struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "INVALID_REQUEST",
			"message": "入力内容が正しくありません",
			"detail":  err.Error(),
		})
		return
	}

	var user models.User
	err := h.db.Where("email = ? AND is_deleted = ?", req.Email, false).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "INVALID_EMAIL_OR_PASSWORD",
			"message": "メールアドレスまたはパスワードが正しくありません",
			"detail":  nil,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "USER_SEARCH_FAILED",
			"message": "ユーザー確認に失敗しました",
			"detail":  err.Error(),
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "INVALID_EMAIL_OR_PASSWORD",
			"message": "メールアドレスまたはパスワードが正しくありません",
			"detail":  nil,
		})
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "TOKEN_GENERATE_FAILED",
			"message": "トークンの発行に失敗しました",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": LoginResponse{
			AccessToken: accessToken,
			User: UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Role:  user.Role,
			},
		},
		"error":   false,
		"code":    "",
		"message": "ログインに成功しました",
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userId")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"userId": userID,
			"email":  email,
			"role":   role,
		},
		"error":   false,
		"code":    "",
		"message": "認証済みユーザー情報を取得しました",
	})
}
