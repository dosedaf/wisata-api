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

type AdminController struct {
	DB *gorm.DB
}

func (ac *AdminController) CreateWisata(c *gin.Context) {
	var input struct {
		models.Wisata
		TagIDs []uint `json:"tagIds"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", nil))
		return
	}

	wisata := input.Wisata

	baseSlug := utils.GenerateSlug(wisata.Name)
	slug := baseSlug
	var count int64
	var i int = 1

	for {
		ac.DB.Model(&models.Wisata{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
		i++
	}
	wisata.Slug = slug

	if err := ac.DB.Create(&wisata).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membuat wisata", nil))
		return
	}

	if len(input.TagIDs) > 0 {
		var tags []models.Tag
		ac.DB.Where("id IN ?", input.TagIDs).Find(&tags)

		if err := ac.DB.Model(&wisata).Association("Tags").Append(&tags); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Wisata berhasil dibuat, tetapi gagal menghubungkan tag", nil))
			return
		}
		wisata.Tags = tags
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Wisata berhasil dibuat", wisata))
}


func (ac *AdminController) GetBookings(c *gin.Context) {
	status := c.Query("status")
	var bookings []models.Booking

	// HAPUS Preload("Schedule")
	query := ac.DB.Preload("User").Preload("Wisata").Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Find(&bookings)
	c.JSON(http.StatusOK, utils.SuccessResponse("Berhasil", bookings))
}

func (ac *AdminController) VerifyPayment(c *gin.Context) {
	bookingID := c.Param("id")
	var booking models.Booking

	if err := ac.DB.First(&booking, bookingID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Booking tidak ditemukan", nil))
		return
	}

	if booking.Status != "WAITING_VERIFICATION" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Status booking tidak valid untuk diverifikasi", nil))
		return
	}

	booking.Status = "ACTIVE"
	booking.QRCode = "QR-" + utils.RandomString(15)

	if err := ac.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal verifikasi pembayaran", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Pembayaran berhasil diverifikasi, tiket aktif", booking))
}

func (ac *AdminController) RejectPayment(c *gin.Context) {
	bookingID := c.Param("id")
	
	tx := ac.DB.Begin()

	var booking models.Booking
	if err := tx.First(&booking, bookingID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Booking tidak ditemukan", nil))
		return
	}

	if booking.Status != "WAITING_VERIFICATION" && booking.Status != "PENDING" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Booking tidak bisa ditolak", nil))
		return
	}

	booking.Status = "REJECTED"
	if err := tx.Save(&booking).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal menolak booking", nil))
		return
	}

	var wisata models.Wisata
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&wisata, booking.WisataID).Error; err == nil {
		wisata.Capacity += booking.TotalTicket
		tx.Save(&wisata)
	}

	tx.Commit()
	c.JSON(http.StatusOK, utils.SuccessResponse("Pembayaran ditolak, kuota tiket dikembalikan", nil))
}

func (ac *AdminController) ScanTicket(c *gin.Context) {
	var input struct {
		QRCode string `json:"qrCode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("QR Code harus diisi", nil))
		return
	}

	var booking models.Booking
	if err := ac.DB.Preload("Wisata").Preload("User").Where("qr_code = ?", input.QRCode).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Tiket tidak ditemukan / QR Code tidak valid", nil))
		return
	}

	if booking.Status == "COMPLETED" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Tiket ini sudah digunakan (Sudah Check-in)", nil))
		return
	}

	if booking.Status != "ACTIVE" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Tiket belum aktif atau belum dibayar", nil))
		return
	}

	if time.Now().After(booking.ValidUntil) {
		booking.Status = "EXPIRED"
		ac.DB.Save(&booking) 
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Mohon maaf, masa berlaku tiket ini sudah habis", nil))
		return
	}

	booking.Status = "COMPLETED"
	if err := ac.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal memproses check-in", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Check-in berhasil! Selamat menikmati wisata", gin.H{
		"wisata":      booking.Wisata.Name,
		"visitorName": booking.User.Name,
		"totalTicket": booking.TotalTicket,
	}))
}

func (ac *AdminController) UpdateWisata(c *gin.Context) {
	id := c.Param("id")
	var wisata models.Wisata

	if err := ac.DB.Preload("Tags").First(&wisata, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Wisata tidak ditemukan", nil))
		return
	}

	var input struct {
		models.Wisata
		TagIDs []uint `json:"tagIds"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validasi gagal", nil))
		return
	}

	if err := ac.DB.Model(&wisata).Updates(input.Wisata).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal memperbarui wisata", nil))
		return
	}

	if input.TagIDs != nil {
		var tags []models.Tag
		if len(input.TagIDs) > 0 {
			ac.DB.Where("id IN ?", input.TagIDs).Find(&tags)
		}
		
		if err := ac.DB.Model(&wisata).Association("Tags").Replace(&tags); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Data wisata terupdate, tapi gagal memperbarui tag", nil))
			return
		}
		wisata.Tags = tags
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Wisata berhasil diperbarui", wisata))
}

func (ac *AdminController) DeleteWisata(c *gin.Context) {
	id := c.Param("id")
	var wisata models.Wisata

	if err := ac.DB.First(&wisata, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Wisata tidak ditemukan", nil))
		return
	}
	
	if err := ac.DB.Model(&wisata).Association("Tags").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membersihkan relasi tag", nil))
		return
	}

	if err := ac.DB.Delete(&wisata).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal menghapus wisata", nil))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Wisata berhasil dihapus", nil))
}
