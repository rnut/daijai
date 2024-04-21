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
func (c *Uploader) DownloadFile(path string) (*storage.Reader, error) {
	ctx := context.Background()
	rc, e := c.Cl.Bucket(c.BucketName).Object(path).NewReader(ctx)
	if e != nil {
		return nil, e
	}
	defer rc.Close()
	return rc, nil
}
