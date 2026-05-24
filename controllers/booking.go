package controllers

import (
	"fmt"
	"net/http"
	"time"
	"wisata-api/models"
	"wisata-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookingController struct {
	DB *gorm.DB
}

func (bc *BookingController) CreateBooking(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input struct {
		WisataID    uint `json:"wisataId" binding:"required"`
		TotalTicket int  `json:"totalTicket" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", err.Error()))
		return
	}

	tx := bc.DB.Begin()

	var wisata models.Wisata
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&wisata, input.WisataID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Wisata tidak ditemukan", nil))
		return
	}

	if wisata.Capacity < input.TotalTicket {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Mohon maaf, sisa slot tiket tidak mencukupi", nil))
		return
	}

	wisata.Capacity -= input.TotalTicket
	if err := tx.Save(&wisata).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal memproses kuota tiket", nil))
		return
	}

	totalPrice := wisata.TicketPrice * float64(input.TotalTicket)
	bookingCode := fmt.Sprintf("BOOK-%d-%d", time.Now().Unix(), userID)
	
	validUntil := time.Now().AddDate(0, 0, 30) 

	booking := models.Booking{
		UserID:      userID.(uint),
		WisataID:    input.WisataID,
		BookingCode: bookingCode,
		TotalTicket: input.TotalTicket,
		TotalPrice:  totalPrice,
		Status:      "PENDING",
		ValidUntil:  validUntil,
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membuat booking", nil))
		return
	}

	tx.Commit() 

	c.JSON(http.StatusCreated, utils.SuccessResponse("Booking berhasil dibuat. Tiket berlaku selama 30 hari.", booking))
}

func (bc *BookingController) CancelBooking(c *gin.Context) {
	userID, _ := c.Get("user_id")
	bookingID := c.Param("id")

	tx := bc.DB.Begin()

	var booking models.Booking
	if err := tx.Where("id = ? AND user_id = ?", bookingID, userID).First(&booking).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Booking tidak ditemukan", nil))
		return
	}

	if booking.Status != "PENDING" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Hanya booking yang belum dibayar (PENDING) yang dapat dibatalkan", nil))
		return
	}

	booking.Status = "CANCELED"
	if err := tx.Save(&booking).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membatalkan booking", nil))
		return
	}

	var wisata models.Wisata
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&wisata, booking.WisataID).Error; err == nil {
		wisata.Capacity += booking.TotalTicket
		tx.Save(&wisata)
	}

	tx.Commit()

	c.JSON(http.StatusOK, utils.SuccessResponse("Pesanan berhasil dibatalkan, stok tiket dikembalikan", nil))
}
