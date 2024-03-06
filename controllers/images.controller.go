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
	prefix := "public"
	directory := c.Param("directory")
	object := c.Param("fileName")
	localDirectory := fmt.Sprintf("%s/%s/%s", prefix, directory, object)

	if _, err := os.Stat(localDirectory); os.IsNotExist(err) {
		uploader.UploadPath = fmt.Sprintf("%s/", directory)
		log.Println("File not found, downloading from cloud storage")
		if object == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "object is required",
			})
			return
		}

		file, err := os.Create(localDirectory)
		if err != nil {
			log.Println("Failed to create file")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		log.Println("Downloading file from cloud storage path: ", object)
		err = uploader.DownloadFile(file, object)
		if err != nil {
			// os.Remove(path)
			log.Println("Failed to download file")
			log.Println("error: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.File(localDirectory)
	} else {
		c.File(localDirectory)
	}
}
