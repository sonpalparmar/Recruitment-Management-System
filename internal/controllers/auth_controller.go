package controllers

import (
	"net/http"

	"github.com/GolangAssignment/internal/config"
	"github.com/GolangAssignment/internal/models"
	"github.com/GolangAssignment/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB  *gorm.DB
	Cfg config.Config
}

func NewAuthController(db *gorm.DB, cfg config.Config) *AuthController {
	return &AuthController{DB: db, Cfg: cfg}
}

type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	UserType        string `json:"user_type" binding:"required,oneof=Admin Applicant"`
	ProfileHeadline string `json:"profile_headline"`
	Address         string `json:"address"`
}

func (ac *AuthController) SignUp(c *gin.Context) {
	var input SignUpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := models.User{
		Name:            input.Name,
		Email:           input.Email,
		Address:         input.Address,
		UserType:        models.UserType(input.UserType),
		PasswordHash:    hashedPassword,
		ProfileHeadline: input.ProfileHeadline,
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Email already exists")
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, gin.H{"message": "User registered successfully"})
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (ac *AuthController) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := ac.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := utils.GenerateToken(user.ID, string(user.UserType), ac.Cfg.JWTSecret)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"token": token})
}
