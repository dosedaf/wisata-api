package main

import (
	"log"
	"wisata-api/controllers"
	"wisata-api/middlewares"
	"wisata-api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "yoda:@tcp(127.0.0.1:3306)/wisata_db?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}

	db.AutoMigrate(
		&models.User{}, &models.Wisata{}, &models.WisataGallery{},
		&models.Tag{}, &models.Schedule{}, &models.Booking{}, &models.Review{},
	)

	authC := &controllers.AuthController{DB: db}
	wisataC := &controllers.WisataController{DB: db}
	bookingC := &controllers.BookingController{DB: db}
	userC := &controllers.UserController{DB: db}
	trxC := &controllers.TransactionController{DB: db}
	adminC := &controllers.AdminController{DB: db}

	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/register", authC.Register)
		auth.POST("/login", authC.Login)
		auth.POST("/logout", func(c *gin.Context) { c.JSON(200, gin.H{"success": true, "message": "Logout berhasil"}) })
	}

	wisata := r.Group("/wisata")
	{
		wisata.GET("/featured", wisataC.GetFeatured)
		wisata.GET("/search", wisataC.Search)
		wisata.GET("/tag/:tag", wisataC.GetByTag)
		wisata.GET("/:slug", wisataC.GetDetail)
		wisata.GET("/:slug/schedules", wisataC.GetSchedules)
	}

	api := r.Group("/")
	api.Use(middlewares.RequireAuth())
	{
		api.GET("/profile", userC.GetProfile)
		api.PUT("/profile", userC.UpdateProfile)

		api.POST("/bookings", bookingC.CreateBooking)
		api.POST("/payments/upload", trxC.UploadPayment)
		
		api.GET("/my-tickets", userC.GetMyTickets)
		api.GET("/history", userC.GetHistory)
		
		api.POST("/reviews", trxC.CreateReview)
	}

	admin := r.Group("/admin")
	admin.Use(middlewares.RequireAuth(), middlewares.RequireAdmin())
	{
		admin.POST("/wisata", adminC.CreateWisata)
		admin.POST("/schedules", adminC.CreateSchedule)
		
		admin.GET("/bookings", adminC.GetBookings)
		admin.PUT("/bookings/:id/verify", adminC.VerifyPayment)
		admin.PUT("/bookings/:id/reject", adminC.RejectPayment)
	}

	r.Run(":8080")
}
