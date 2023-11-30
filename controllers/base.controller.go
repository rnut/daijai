package controllers

import (
	"daijai/models"
	"daijai/token"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseController struct {
	DB *gorm.DB
}

func (bc *BaseController) GetUserID(c *gin.Context, userID *uint) error {
	uid, err := token.ExtractTokenID(c)
	if err != nil {
		return err
	}
	*userID = uid
	return nil
}

// Implement this method in your controller to fetch user data
func (bc *BaseController) getUserDataByUserID(db *gorm.DB, userID uint, member *models.Member) error {
	var u models.User
	if err := db.First(&u, userID).Error; err != nil {
		log.Print(err)
		return err
	}
	*member = u.UserToMember()
	return nil
}
