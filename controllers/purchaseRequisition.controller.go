package controllers

import (
	"daijai/models"
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
	var orderBoms []models.OrderBom
	if err := prc.DB.
		Preload("Order").
		Preload("Bom.Material.Category").
		Where("is_full_filled = ?", false).
		Find(&orderBoms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orderBoms"})
		return
	}

	var resp struct {
		OrderBoms []models.OrderBom
		Slug	  string
	}

	resp.OrderBoms = orderBoms
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
		Slug string `json:"Slug"`
		Notes string `json:"Notes"`
		PurchaseMaterials []models.PurchaseMaterial `json:"PurchaseMaterials"`
		PORefs []string `json:"PORefs"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	var poRefs []models.PORef
	for _, slug := range request.PORefs {
		po := models.PORef {
			Slug: slug,
		}
		if err := prc.DB.
			Where("slug = ?", slug).
			FirstOrCreate(&po).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create poRef"})
			return	
		}

		poRefs = append(poRefs, po)
	}

	// create Purchase
	purchase := models.Purchase{
		Slug: request.Slug,
		Notes: request.Notes,
		CreatedByID: member.ID,
		PORefs: poRefs,
	}
	if err := prc.DB.Create(&purchase).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H {
		"purchase": purchase,
	})


	// for i, rm := range request.PurchaseMaterials {
	// 	material := models.Material{}
	// 	if err := prc.DB.First(&material, rm.MaterialID).Error; err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// 	if err := prc.DB.Save(&material).Error; err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// 	if err := prc.DB.Save(&request.PurchaseMaterials[i]).Error; err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// }

	

	// c.JSON(http.StatusOK, purchase)

	// request.CreatedByID = member.ID
	// if err := prc.DB.Transaction(func(tx *gorm.DB) error {
	// 	if err := tx.Create(&request).Error; err != nil {
	// 		return err
	// 	}

	// 	// Update the associated materials' quantity
	// 	for i, rm := range request.PurchaseMaterials {
	// 		material := models.Material{}
	// 		if err := prc.DB.First(&material, rm.MaterialID).Error; err != nil {
	// 			return err
	// 		}

	// 		// Update the material's quantity
	// 		// material.IncomingQuantity += rm.Quantity

	// 		if err := tx.Save(&material).Error; err != nil {
	// 			tx.Rollback()
	// 			return err
	// 		}

	// 		// set material back to withdrawalMaterials
	// 		request.PurchaseMaterials[i].Material = material

	// 		if err := tx.Save(&request.PurchaseMaterials[i]).Error; err != nil {
	// 			return err
	// 		}
	// 	}

	// 	return nil
	// }); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pr"})
	// 	return
	// }

	// c.JSON(http.StatusCreated, gin.H{"message": "PurchaseRequisition created successfully"})
}

// GetPurchaseRequisition retrieves a PurchaseRequisition by ID.
func (prc *PurchaseRequisitionController) GetPurchaseRequisition(c *gin.Context) {
	objID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Withdrawal ID"})
		return
	}

	var purchaseRequisition models.Purchase
	if err := prc.
		DB.
		Preload("PurchaseMaterials.Material.Category").
		Preload("CreatedBy").
		Preload("Project").
		First(&purchaseRequisition, objID).
		Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PurchaseRequisition not found"})
		return
	}

	c.JSON(http.StatusOK, purchaseRequisition)
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
		Preload("CreatedBy").
		Preload("Project")
	if member.Role == "admin" {
		q.Find(&ps)
	} else {
		q.Find(&ps, "created_by_id = ?", member.ID)
	}
	if err := q.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
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
