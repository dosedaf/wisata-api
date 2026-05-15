package controllers

import (
	"net/http"
	"wisata-api/models"
	"wisata-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB *gorm.DB
}

func (ac *AdminController) CreateWisata(c *gin.Context) {
	var input models.Wisata

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", nil))
		return
	}

	if err := ac.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membuat wisata", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Wisata berhasil dibuat", nil))
}

func (ac *AdminController) CreateSchedule(c *gin.Context) {
	var input models.Schedule

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", nil))
		return
	}

	input.RemainingQuota = input.Quota

	if err := ac.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membuat jadwal", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Schedule berhasil dibuat", nil))
}
