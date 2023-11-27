package controllers

import (
	"daijai/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProjectController handles CRUD operations for Project.
type ProjectController struct {
	DB *gorm.DB
	BaseController
}

func NewProjectController(db *gorm.DB) *ProjectController {
	return &ProjectController{
		DB: db,
	}
}

// CreateProject handles the creation of a new Project.
func (pc *ProjectController) CreateProject(c *gin.Context) {
	var request models.Project

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new Project
	if err := pc.DB.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Project"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Project created successfully", "project": request})
}

func (pc *ProjectController) GetAllProjects(c *gin.Context) {
	uid, err := pc.GetUserID(c)
	if err != nil {
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"uid": uid,
		})
		return
	}

	var projects []models.Project
	if err := pc.DB.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject retrieves a Project by ID.
func (pc *ProjectController) GetProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := pc.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates a Project by ID.
func (pc *ProjectController) UpdateProject(c *gin.Context) {
	id := c.Param("id")

	var request models.Project
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := pc.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Update the Project fields
	project.Slug = request.Slug
	project.Title = request.Title
	project.Subtitle = request.Subtitle
	project.Description = request.Description

	if err := pc.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project updated successfully", "project": project})
}

// DeleteProject deletes a Project by ID.
func (pc *ProjectController) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := pc.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if err := pc.DB.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}
