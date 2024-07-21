package controllers

import (
	"daijai/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectStoreController struct {
	DB *gorm.DB
	BaseController
}

func NewProjectStoreController(db *gorm.DB) *ProjectStoreController {
	return &ProjectStoreController{DB: db}
}

// create project store
func (p *ProjectStoreController) CreateProjectStore(c *gin.Context) {
	projectStore := models.ProjectStore{}
	if err := c.ShouldBindJSON(&projectStore); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := p.DB.Create(&projectStore)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, projectStore)
}

// get all project stores
func (p *ProjectStoreController) GetProjectStores(c *gin.Context) {
	projectStores := []models.ProjectStore{}
	result := p.DB.Find(&projectStores)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, projectStores)
}

// get project store by id
func (p *ProjectStoreController) GetProjectStoreByID(c *gin.Context) {
	projectStore := models.ProjectStore{}
	id := c.Param("id")
	result := p.DB.First(&projectStore, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, projectStore)
}

// update project store
func (p *ProjectStoreController) UpdateProjectStore(c *gin.Context) {
	projectStore := models.ProjectStore{}
	id := c.Param("id")
	result := p.DB.First(&projectStore, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	if err := c.ShouldBindJSON(&projectStore); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result = p.DB.Save(&projectStore)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, projectStore)
}

// delete project store
func (p *ProjectStoreController) DeleteProjectStore(c *gin.Context) {
	projectStore := models.ProjectStore{}
	id := c.Param("id")
	result := p.DB.First(&projectStore, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	result = p.DB.Delete(&projectStore)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project Store deleted successfully"})
}
