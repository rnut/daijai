package controllers

import (
	"daijai/models"
	"daijai/token"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
	BaseController
}

func NewAuth(db *gorm.DB) *AuthController {
	return &AuthController{
		DB: db,
	}
}

func (uc *AuthController) Session(c *gin.Context) {
	var uid uint
	if err := uc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := uc.getUserDataByUserID(uc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, member)
}

// Register creates a new user.
func (uc *AuthController) Register(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashedPassword)
	if err := uc.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})
}

// Login authenticates a user and generates a JWT token.
func (uc *AuthController) Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var user models.User
	if err := uc.DB.Where("username = ?", loginData.Username).First(&user).Error; err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := token.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error creating token"})
		return
	}

	// tk := models.Tokens{
	// 	AccessToken: tokenString,
	// }
	// data := models.ResponseToken{
	// 	Tokens: models.Tokens{
	// 	},
	// }
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

// Logout revokes the user's JWT token (optional depending on your requirements).
func (uc *AuthController) Logout(c *gin.Context) {
	// Perform any necessary logout actions, if applicable (optional)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
