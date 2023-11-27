package controllers

import (
	"daijai/models"
	"daijai/token"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseController struct {
	DB *gorm.DB
}

func (bc *BaseController) GetUserID(c *gin.Context) (uint, error) {
	uid, err := token.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id"})
		return 0, err
	}
	return uid, nil
}

// Implement this method in your controller to fetch user data
func (bc *BaseController) getUserDataByUserID(userID uint) *models.User {
	var user models.User
	if err := bc.DB.First(&user, userID).Error; err != nil {
		return &user
	}
	return nil
}
