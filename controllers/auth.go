package controllers

import (
	"net/http"
	"wisata-api/models"
	"wisata-api/utils"
	"wisata-api/middlewares"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func (ac *AuthController) Register(c *gin.Context) {
	var input struct {
		Name            string `json:"name" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		Password        string `json:"password" binding:"required,min=6"`
		ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", err.Error()))
		return
	}

	var existingUser models.User
	if ac.DB.Where("email = ?", input.Email).First(&existingUser).RowsAffected > 0 {
		c.JSON(http.StatusConflict, utils.ErrorResponse("Email sudah digunakan", map[string]string{"email": "Email already exists"}))
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     "USER",
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Terjadi kesalahan sistem", nil))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse("Register berhasil", user))
}

func (ac *AuthController) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Format input salah", nil))
		return
	}

	var user models.User
	if err := ac.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Email atau password salah", nil))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Email atau password salah", nil))
		return
	}

	token, _ := middlewares.GenerateToken(user.ID, user.Role)

	c.JSON(http.StatusOK, utils.SuccessResponse("Login berhasil", gin.H{
		"token": token,
		"user":  user,
	}))
}
