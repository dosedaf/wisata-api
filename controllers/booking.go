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
		ScheduleID  uint `json:"scheduleId" binding:"required"`
		TotalTicket int  `json:"totalTicket" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", err.Error()))
		return
	}

	tx := bc.DB.Begin()

	var wisata models.Wisata
	if err := tx.First(&wisata, input.WisataID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Wisata tidak ditemukan", nil))
		return
	}

	var schedule models.Schedule
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&schedule, input.ScheduleID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Jadwal tidak ditemukan", nil))
		return
	}

	if schedule.RemainingQuota < input.TotalTicket {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Kuota tiket tidak mencukupi", nil))
		return
	}

	schedule.RemainingQuota -= input.TotalTicket
	if err := tx.Save(&schedule).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal memproses kuota", nil))
		return
	}

	totalPrice := wisata.TicketPrice * float64(input.TotalTicket)
	bookingCode := fmt.Sprintf("BOOK-%d-%d", time.Now().Unix(), userID)

	booking := models.Booking{
		UserID:      userID.(uint),
		WisataID:    input.WisataID,
		ScheduleID:  input.ScheduleID,
		BookingCode: bookingCode,
		TotalTicket: input.TotalTicket,
		TotalPrice:  totalPrice,
		Status:      "PENDING",
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membuat booking", nil))
		return
	}

	tx.Commit() 

	c.JSON(http.StatusCreated, utils.SuccessResponse("Booking berhasil dibuat", booking))
}
