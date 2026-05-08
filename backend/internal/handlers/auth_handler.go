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

type MeResponse struct {
	UserID uint   `json:"userId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
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

/*
 * ログイン中ユーザー情報取得
 *
 * AuthMiddlewareでJWTを検証済み。
 * JWTから取得したuserIdを使ってDBから最新のユーザー情報を取得する。
 *
 * name / email / role はDBの最新情報を返す。
 */
func (h *AuthHandler) Me(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "USER_ID_NOT_FOUND",
			"message": "認証情報からユーザーIDを取得できません",
			"detail":  nil,
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "INVALID_USER_ID",
			"message": "認証情報のユーザーIDが正しくありません",
			"detail":  nil,
		})
		return
	}

	var user models.User
	err := h.db.Where("id = ? AND is_deleted = ?", userID, false).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "USER_NOT_FOUND",
			"message": "ログイン中のユーザーが存在しません",
			"detail":  nil,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "USER_SEARCH_FAILED",
			"message": "ログイン中ユーザー情報の取得に失敗しました",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": MeResponse{
			UserID: user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Role:   user.Role,
		},
		"error":   false,
		"code":    "",
		"message": "認証済みユーザー情報を取得しました",
	})
}
