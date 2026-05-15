package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"wisata-api/utils"
	"github.com/gin-gonic/gin"
)

type UploadController struct{}

func (uc *UploadController) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Gagal mengambil file", err.Error()))
		return
	}

	uploadDir := "public/uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	savePath := fmt.Sprintf("%s/%s", uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal menyimpan file", err.Error()))
		return
	}

	fileURL := fmt.Sprintf("http://localhost:8080/%s", savePath)

	c.JSON(http.StatusOK, utils.SuccessResponse("File berhasil diunggah", gin.H{
		"url": fileURL,
	}))
}
