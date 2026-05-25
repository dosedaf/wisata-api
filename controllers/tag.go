package controllers

import (
	"encoding/json"
	"net/http"
	"time"
	"wisata-api/models"
	"wisata-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagController struct {
	DB    *gorm.DB
	Cache *utils.MemoryCache 
}

func (tc *TagController) GetAllTags(c *gin.Context) {
	cacheKey := "wisata:all_tags"

	if cachedData, found := tc.Cache.Get(cacheKey); found {
		var tags []models.Tag
		json.Unmarshal(cachedData, &tags)
		
		c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil mengambil data tag (dari Cache)", tags))
		return
	}

	var tags []models.Tag
	tc.DB.Find(&tags)

	tagsJSON, err := json.Marshal(tags)
	if err == nil {
		tc.Cache.Set(cacheKey, tagsJSON, 24*time.Hour)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil mengambil data tag (dari MySQL)", tags))
}
