package controllers

import (
	"net/http"
	"wisata-api/models"
	"wisata-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func (uc *UserController) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User

	if err := uc.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("User tidak ditemukan", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", user))
}

func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phoneNumber"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Format input salah", nil))
		return
	}

	uc.DB.Model(&models.User{}).Where("id = ?", userID).Updates(input)
	c.JSON(http.StatusOK, utils.SuccessResponse("Profile berhasil diperbarui", nil))
}

func (uc *UserController) GetAllTickets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var tickets []models.Booking

	uc.DB.Preload("Wisata").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tickets)

	response := []map[string]interface{}{}
	for _, t := range tickets {
		response = append(response, map[string]interface{}{
			"bookingId":   t.ID,
			"bookingCode": t.BookingCode,
			"status":      t.Status,
			"validUntil":  t.ValidUntil,
			"totalTicket": t.TotalTicket,
			"qrCode":      t.QRCode,
			"wisata": map[string]interface{}{
				"name":     t.Wisata.Name,
				"imageUrl": t.Wisata.ImageUrl,
			},
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", response))
}

func (uc *UserController) GetActiveTickets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var tickets []models.Booking

	uc.DB.Preload("Wisata").
		Where("user_id = ? AND status = ?", userID, "ACTIVE").
		Order("created_at DESC").
		Find(&tickets)

	response := []map[string]interface{}{}
	for _, t := range tickets {
		response = append(response, map[string]interface{}{
			"bookingId":   t.ID,
			"bookingCode": t.BookingCode,
			"status":      t.Status,
			"validUntil":  t.ValidUntil,
			"totalTicket": t.TotalTicket,
			"qrCode":      t.QRCode,
			"wisata": map[string]interface{}{
				"name":     t.Wisata.Name,
				"imageUrl": t.Wisata.ImageUrl,
			},
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", response))
}

func (uc *UserController) GetPendingTickets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var tickets []models.Booking

	uc.DB.Preload("Wisata").
		Where("user_id = ? AND status IN ?", userID, []string{"PENDING", "WAITING_VERIFICATION"}).
		Order("created_at DESC").
		Find(&tickets)

	response := []map[string]interface{}{}
	for _, t := range tickets {
		response = append(response, map[string]interface{}{
			"bookingId":   t.ID,
			"bookingCode": t.BookingCode,
			"status":      t.Status,
			"validUntil":  t.ValidUntil,
			"totalTicket": t.TotalTicket,
			"qrCode":      t.QRCode,
			"wisata": map[string]interface{}{
				"name":     t.Wisata.Name,
				"imageUrl": t.Wisata.ImageUrl,
			},
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", response))
}

func (uc *UserController) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var histories []models.Booking

	uc.DB.Preload("Wisata").
		Where("user_id = ? AND status = ?", userID, "COMPLETED").
		Order("created_at DESC").
		Find(&histories)

	response := []map[string]interface{}{}
	for _, h := range histories {
		response = append(response, map[string]interface{}{
			"bookingId":   h.ID,
			"validUntil":  h.ValidUntil, 
			"status":      h.Status,
			"hasReviewed": h.HasReviewed,
			"wisata": map[string]interface{}{
				"id":       h.Wisata.ID,
				"name":     h.Wisata.Name,
				"imageUrl": h.Wisata.ImageUrl,
			},
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", response))
}
