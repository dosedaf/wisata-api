package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"
	"wisata-api/utils"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

type UploadController struct{}

func (uc *UploadController) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Gagal mengambil file", err.Error()))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membuka file", err.Error()))
		return
	}
	defer f.Close()

	ctx := context.Background()
	
	client, err := storage.NewClient(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal koneksi ke Cloud Storage", err.Error()))
		return
	}
	defer client.Close()

	bucketName := "wisata-api-bucket" 
	
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))

	wc := client.Bucket(bucketName).Object(filename).NewWriter(ctx)
	if _, err := io.Copy(wc, f); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal mengunggah ke Cloud Storage", err.Error()))
		return
	}
	
	if err := wc.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal menyelesaikan unggahan", err.Error()))
		return
	}

	fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filename)

	c.JSON(http.StatusOK, utils.SuccessResponse("File berhasil diunggah", gin.H{
		"url": fileURL,
	}))
}
