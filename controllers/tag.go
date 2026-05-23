package controllers

import (
	"net/http"
	"wisata-api/models"
	"wisata-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagController struct {
	DB *gorm.DB
}

func (tc *TagController) GetAllTags(c *gin.Context) {
	var tags []models.Tag
	tc.DB.Find(&tags)

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil mengambil data tag", tags))
}
