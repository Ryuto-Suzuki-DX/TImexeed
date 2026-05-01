package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/auth"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

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

	var existingUser models.User
	err := h.db.Where("email = ? AND is_deleted = ?", req.Email, false).First(&existingUser).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "EMAIL_ALREADY_EXISTS",
			"message": "このメールアドレスはすでに登録されています",
			"detail":  nil,
		})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "USER_SEARCH_FAILED",
			"message": "ユーザー確認に失敗しました",
			"detail":  err.Error(),
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "PASSWORD_HASH_FAILED",
			"message": "パスワードの処理に失敗しました",
			"detail":  err.Error(),
		})
		return
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "USER",
		IsDeleted:    false,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"error":   true,
			"code":    "USER_CREATE_FAILED",
			"message": "ユーザー登録に失敗しました",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
		"error":   false,
		"code":    "",
		"message": "ユーザー登録が完了しました",
	})
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