package controllers

import (
	"daijai/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PurchaseRequisitionController handles CRUD operations for PurchaseRequisition.
type PurchaseRequisitionController struct {
	DB *gorm.DB
	BaseController
}

func NewPurchaseRequisitionController(db *gorm.DB) *PurchaseRequisitionController {
	return &PurchaseRequisitionController{
		DB: db,
	}
}

// get list or orderBoms with IsFullFilled = false
func (prc *PurchaseRequisitionController) GetNewPRInfo(c *gin.Context) {
	var purchaseSuggestions []models.PurchaseSuggestion
	if err := prc.DB.
		Preload("OrderBom.Order.Drawing").
		Preload("OrderBom.Bom.Material").
		Where("status IN (?)", []string{models.PurchaseSuggestionStatus_Ready, models.PurchaseSuggestionStatus_InProgress}).
		Find(&purchaseSuggestions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve purchaseSuggestions"})
		return
	}

	var poRefs []models.PORef
	if err := prc.DB.Find(&poRefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve poRefs"})
		return
	}

	var categories []models.Category
	if err := prc.DB.Preload("Materials").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	var resp struct {
		Categories          []models.Category
		PurchaseSuggestions []models.PurchaseSuggestion
		Slug                string
		PORefs              []models.PORef
	}

	resp.Categories = categories
	resp.PurchaseSuggestions = purchaseSuggestions
	resp.PORefs = poRefs
	if err := prc.RequestSlug(&resp.Slug, prc.DB, "withdrawals"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slug", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreatePurchaseRequisition handles the creation of a new PurchaseRequisition.
func (prc *PurchaseRequisitionController) CreatePurchaseRequisition(c *gin.Context) {
	var uid uint
	if err := prc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := prc.getUserDataByUserID(prc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var request struct {
		Slug              string                    `json:"Slug"`
		Notes             string                    `json:"Notes"`
		PurchaseMaterials []models.PurchaseMaterial `json:"PurchaseMaterials"`
		PORefs            []models.PORef            `json:"PORefs"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := prc.DB.Transaction(func(tx *gorm.DB) error {
		// create Purchase
		purchase := models.Purchase{
			Slug:        request.Slug,
			Notes:       request.Notes,
			CreatedByID: member.ID,
			PORefs:      request.PORefs,
		}
		if err := tx.Create(&purchase).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return err
		}

		// create purchaseMaterials
		for _, v := range request.PurchaseMaterials {
			prm := models.PurchaseMaterial{
				PurchaseID: purchase.ID,
				MaterialID: v.MaterialID,
				Quantity:   v.Quantity,
			}
			if err := tx.Create(&prm).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Order"})
		log.Println(err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "PurchaseRequisition created successfully"})
}

// GetPurchaseRequisition retrieves a PurchaseRequisition by Slug.
func (prc *PurchaseRequisitionController) GetPurchaseRequisition(c *gin.Context) {
	slug := c.Param("slug")

	var purchaseRequisition models.Purchase
	if err := prc.
		DB.
		Preload("PurchaseMaterials.Material").
		Preload("PORefs").
		Preload("CreatedBy").
		First(&purchaseRequisition, "slug = ?", slug).
		Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PurchaseRequisition not found"})
		return
	}

	var poRefs []string
	for _, v := range purchaseRequisition.PORefs {
		poRefs = append(poRefs, v.Slug)
	}

	var ivtMats []models.InventoryMaterial
	if err := prc.
		DB.
		Joins("Receipt").
		Where("po_number IN (?)", poRefs).
		Find(&ivtMats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory material transactions"})
		return
	}

	var ivtIDs []uint
	for _, ivtMat := range ivtMats {
		ivtIDs = append(ivtIDs, ivtMat.ID)
	}

	var transactions []models.InventoryMaterialTransaction
	if err := prc.
		DB.
		Where("inventory_material_id IN ?", ivtIDs).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory material transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"purchaseRequisition": purchaseRequisition,
		"transactions":        transactions,
	})
}

func (pc *PurchaseRequisitionController) GetAllPurchaseRequisition(c *gin.Context) {
	var uid uint
	if err := pc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := pc.getUserDataByUserID(pc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var ps []models.Purchase
	q := pc.DB.
		Preload("PurchaseMaterials.Material.Category").
		Preload("CreatedBy")
	if member.Role == "admin" {
		q.Find(&ps)
	} else {
		q.Find(&ps, "created_by_id = ?", member.ID)
	}
	if err := q.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve purchase requisitions"})
		return
	}
	c.JSON(http.StatusOK, ps)
}

// UpdatePurchaseRequisition updates a PurchaseRequisition by ID.
func (prc *PurchaseRequisitionController) UpdatePurchaseRequisition(c *gin.Context) {
	var request models.Purchase
	prID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawing ID"})
		return
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var purchaseRequisition models.Purchase
	if err := prc.DB.Preload("PurchaseMaterials.Material.Category").First(&purchaseRequisition, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PurchaseRequisition not found"})
		return
	}

	purchaseRequisition.Notes = request.Notes

	if err := prc.DB.Transaction(func(tx *gorm.DB) error {
		if err := prc.DB.Save(&purchaseRequisition).Error; err != nil {
			return err
		}

		// DELETE ALL WithdrawalMaterials
		for _, v := range purchaseRequisition.PurchaseMaterials {
			if err := tx.Delete(&models.PurchaseMaterial{}, v.ID).Error; err != nil {
				return err
			}
		}

		for _, v := range request.PurchaseMaterials {
			wm := models.PurchaseMaterial{
				PurchaseID: purchaseRequisition.ID,
				MaterialID: v.MaterialID,
				Quantity:   v.Quantity,
			}
			if err := tx.Create(&wm).Error; err != nil {
				return err
			}
			purchaseRequisition.PurchaseMaterials = append(purchaseRequisition.PurchaseMaterials, wm)
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PurchaseRequisition updated successfully", "purchaseRequisition": purchaseRequisition})
}

// DeletePurchaseRequisition deletes a PurchaseRequisition by ID.
func (prc *PurchaseRequisitionController) DeletePurchaseRequisition(c *gin.Context) {
	id := c.Param("id")

	var purchaseRequisition models.Purchase
	if err := prc.DB.First(&purchaseRequisition, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PurchaseRequisition not found"})
		return
	}

	if err := prc.DB.Delete(&purchaseRequisition).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete PurchaseRequisition"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PurchaseRequisition deleted successfully"})
}
