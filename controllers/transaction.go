package controllers

import (
	"net/http"
	"wisata-api/models"
	"wisata-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionController struct {
	DB *gorm.DB
}

func (tc *TransactionController) CreateReview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		BookingID uint   `json:"bookingId" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Comment   string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", nil))
		return
	}

	var booking models.Booking
	if err := tc.DB.Where("id = ? AND user_id = ?", input.BookingID, userID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Booking tidak ditemukan", nil))
		return
	}

	if booking.Status != "COMPLETED" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Booking belum selesai", nil))
		return
	}

	if booking.HasReviewed {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Anda sudah memberikan review", nil))
		return
	}

	review := models.Review{
		WisataID:  booking.WisataID,
		UserID:    userID.(uint),
		BookingID: booking.ID,
		Rating:    input.Rating,
		Comment:   input.Comment,
	}

	tx := tc.DB.Begin()
	tx.Create(&review)
	tx.Model(&booking).Update("has_reviewed", true)
	tx.Commit()

	c.JSON(http.StatusOK, utils.SuccessResponse("Review berhasil ditambahkan", review))
}

func (tc *TransactionController) UploadPayment(c *gin.Context) {
	var input struct {
		BookingID     uint   `json:"bookingId" binding:"required"`
		PaymentMethod string `json:"paymentMethod" binding:"required"`
		PaymentProof  string `json:"paymentProof" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Input tidak valid", nil))
		return
	}

	// Ubah status menjadi menunggu verifikasi
	if err := tc.DB.Model(&models.Booking{}).Where("id = ? AND status = ?", input.BookingID, "PENDING").
		Updates(map[string]interface{}{
			"status":         "WAITING_VERIFICATION",
			"payment_method": input.PaymentMethod,
			"payment_proof":  input.PaymentProof,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal memproses pembayaran atau booking sudah diproses", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Pembayaran berhasil dikirim, menunggu verifikasi Admin", nil))
}
