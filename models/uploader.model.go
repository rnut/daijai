package models

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"cloud.google.com/go/storage"
)

type Uploader struct {
	Cl         *storage.Client
	BucketName string
	UploadPath string
}

// UploadFile uploads an object
func (c *Uploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.Cl.Bucket(c.BucketName).Object(c.UploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

// Download file from GCS
func (c *Uploader) DownloadFile(w io.Writer, object string) error {
	ctx := context.Background()
	rc, err := c.Cl.Bucket(c.BucketName).Object(c.UploadPath + object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()

	if _, err := io.Copy(w, rc); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	return nil
}
