package controllers

import (
	"context"
	"daijai/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

type ImageController struct {
}

const (
	bucketName = "daijai-image-bucket"
)

var uploader *models.Uploader

func NewImageController() *ImageController {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "keys/daijai-d4ab4aa6981d.json")
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uploader = &models.Uploader{
		Cl:         client,
		BucketName: bucketName,
		UploadPath: "",
	}

	return &ImageController{}
}

func (ctrl *ImageController) Upload(c *gin.Context) {
	d := c.Param("directory")
	uploader.UploadPath = fmt.Sprintf("%s/", d)

	f, err := c.FormFile("file_input")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	blobFile, err := f.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// get current date time in string
	now := time.Now()
	// format date time to string
	dateTime := now.Format("20060102150405")
	// create file name
	fileName := fmt.Sprintf("%s_%s_%s", d, dateTime, f.Filename)

	err = uploader.UploadFile(blobFile, fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"path":    fmt.Sprintf("/%s/%s", d, fileName),
	})
}

func (ctrl *ImageController) Download(c *gin.Context) {
	directory := c.Param("directory")
	fileName := c.Param("fileName")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "object is required",
		})
		return
	}
	path := fmt.Sprintf("%s/%s", directory, fileName)
	reader, err := uploader.DownloadFile(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.DataFromReader(http.StatusOK, reader.Attrs.Size, "image/jpeg", reader, map[string]string{})
}
