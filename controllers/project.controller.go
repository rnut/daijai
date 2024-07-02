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
	// uid, err := pc.GetUserID(c)
	// if err != nil {
	// 	return
	// } else {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"uid": uid,
	// 	})
	// 	return
	// }

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

// GetProjectDetailBySlug retrieves a Project by Slug.
func (pc *ProjectController) GetProjectDetailBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var response struct {
		Project                       models.Project                        `json:"project"`
		Orders                        []models.Order                        `json:"orders"`
		ExtendOrders                  []models.ExtendOrder                  `json:"extendOrders"`
		InventoryMaterialTransactions []models.InventoryMaterialTransaction `json:"transactions"`
	}
	if err := pc.
		DB.
		Where("slug = ?", slug).
		First(&response.Project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if err := pc.
		DB.
		Preload("Drawing").
		Preload("CreatedBy").
		Where("project_id = ?", response.Project.ID).
		Find(&response.Orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Orders"})
		return
	}

	if err := pc.
		DB.
		Preload("CreatedBy").
		Where("project_id = ?", response.Project.ID).
		Find(&response.ExtendOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Extend Orders"})
		return
	}

	var orderIDs []int64
	var extendOrderIDs []int64

	for _, order := range response.Orders {
		orderIDs = append(orderIDs, int64(order.ID))
	}

	for _, extendOrder := range response.ExtendOrders {
		extendOrderIDs = append(extendOrderIDs, int64(extendOrder.ID))
	}

	if err := pc.
		DB.
		Preload("InventoryMaterial.Material").
		Where("order_id IN (?)", orderIDs).
		Or("extend_order_id IN (?)", extendOrderIDs).
		Find(&response.InventoryMaterialTransactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Inventory Material Transactions"})
		return
	}

	c.JSON(http.StatusOK, response)
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
