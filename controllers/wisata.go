package controllers

import (
	"math"
	"net/http"
	"strconv"
	"wisata-api/models"
	"wisata-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WisataController struct {
	DB *gorm.DB
}

func (wc *WisataController) GetAll(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit

	var wisata []models.Wisata
	var total int64

	wc.DB.Model(&models.Wisata{}).Count(&total)

	wc.DB.Preload("Tags").Limit(limit).Offset(offset).Find(&wisata)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	if wisata == nil {
		wisata = []models.Wisata{}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil mengambil semua data wisata", gin.H{
		"currentPage": page,
		"totalPages":  totalPages,
		"totalData":   total,
		"items":       wisata,
	}))
}

func (wc *WisataController) GetFeatured(c *gin.Context) {
	var wisata []models.Wisata
	wc.DB.Preload("Tags").Order("rating DESC").Limit(5).Find(&wisata)
	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", wisata))
}

func (wc *WisataController) Search(c *gin.Context) {
	q := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 10
	offset := (page - 1) * limit

	var wisata []models.Wisata
	var total int64

	query := wc.DB.Model(&models.Wisata{}).Where("name LIKE ?", "%"+q+"%")
	query.Count(&total)
	query.Preload("Tags").Limit(limit).Offset(offset).Find(&wisata)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", gin.H{
		"currentPage": page,
		"totalPages":  totalPages,
		"totalData":   total,
		"items":       wisata,
	}))
}

func (wc *WisataController) GetByTag(c *gin.Context) {
	tagName := c.Param("tag")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 10
	offset := (page - 1) * limit

	var wisata []models.Wisata
	var total int64
	
	query := wc.DB.Model(&models.Wisata{}).
		Joins("JOIN wisata_tags ON wisata_tags.wisata_id = wisatas.id").
		Joins("JOIN tags ON tags.id = wisata_tags.tag_id").
		Where("tags.name = ?", tagName)

	query.Count(&total)
	query.Preload("Tags").Limit(limit).Offset(offset).Find(&wisata)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", gin.H{
		"currentPage": page,
		"totalPages":  totalPages,
		"totalData":   total,
		"items":       wisata,
	}))
}

func (wc *WisataController) GetDetail(c *gin.Context) {
	slug := c.Param("slug")
	var wisata models.Wisata

	if err := wc.DB.Preload("Galleries").Preload("Tags").Preload("Reviews.User").
		Where("slug = ?", slug).First(&wisata).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Wisata tidak ditemukan", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", wisata))
}

func (wc *WisataController) GetSchedules(c *gin.Context) {
	wisataID := c.Param("id")
	var schedules []models.Schedule

	wc.DB.Where("wisata_id = ?", wisataID).Find(&schedules)

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", schedules))
}
