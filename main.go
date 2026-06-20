package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"wisata-api/controllers"
	"wisata-api/middlewares"
	"wisata-api/models"
	"wisata-api/utils"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func accessSecret() string {
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/991771824221/secrets/new_mysql/versions/latest",
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	return string(result.Payload.Data)
}

func seedTags(db *gorm.DB) {
	hardcodedTags := []string{
		"wisata-air",
		"wisata-sungai",
		"wisata-alam",
		"wisata-budaya",
		"wisata-kuliner",
		"wisata-keluarga",
	}

	for _, tagName := range hardcodedTags {
		var tag models.Tag
		db.Where(models.Tag{Name: tagName}).FirstOrCreate(&tag)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")

	var dbPass string

	env := os.Getenv("ENV")

	if env == "production" {
		dbPass = accessSecret()
	} else {
		dbPass = os.Getenv("DB_PASSWORD")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}

	db.AutoMigrate(
		&models.User{},
		&models.Wisata{},
		&models.WisataGallery{},
		&models.Tag{},
		&models.Booking{},
		&models.Review{},
	)

	seedTags(db)

	memoryCache := utils.NewMemoryCache()

	authC := &controllers.AuthController{DB: db}
	wisataC := &controllers.WisataController{DB: db}
	bookingC := &controllers.BookingController{DB: db}
	userC := &controllers.UserController{DB: db}
	trxC := &controllers.TransactionController{DB: db}
	adminC := &controllers.AdminController{DB: db}
	uploadC := &controllers.UploadController{}

	tagC := &controllers.TagController{DB: db, Cache: memoryCache}

	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true 
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	r.Static("/public", "./public")
	r.GET("/tags", tagC.GetAllTags)

	auth := r.Group("/auth")
	{
		auth.POST("/register", authC.Register)
		auth.POST("/login", authC.Login)
		auth.POST("/logout", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "Logout berhasil",
			})
		})
	}

	wisata := r.Group("/wisata")
	{
		wisata.GET("", wisataC.GetAll)
		wisata.GET("/featured", wisataC.GetFeatured)
		wisata.GET("/search", wisataC.Search)
		wisata.GET("/tag/:tag", wisataC.GetByTag)
		wisata.GET("/:slug", wisataC.GetDetail)
	}

	api := r.Group("/")
	api.Use(middlewares.RequireAuth())
	{
		api.GET("/profile", userC.GetProfile)
		api.PUT("/profile", userC.UpdateProfile)

		api.POST("/bookings", bookingC.CreateBooking)
		api.POST("/payments/upload", trxC.UploadPayment)

		api.GET("/my-tickets", userC.GetAllTickets)
		api.GET("/my-tickets/active", userC.GetActiveTickets)
		api.GET("/my-tickets/pending", userC.GetPendingTickets)
		api.GET("/history", userC.GetHistory)

		api.POST("/reviews", trxC.CreateReview)
		api.POST("/upload", uploadC.UploadFile)
		api.POST("/bookings/:id/cancel", bookingC.CancelBooking)
	}

	admin := r.Group("/admin")
	admin.Use(
		middlewares.RequireAuth(),
		middlewares.RequireAdmin(),
	)
	{
		admin.POST("/wisata", adminC.CreateWisata)

		admin.GET("/bookings", adminC.GetBookings)
		admin.PUT("/bookings/:id/verify", adminC.VerifyPayment)
		admin.PUT("/bookings/:id/reject", adminC.RejectPayment)
		admin.POST("/bookings/scan", adminC.ScanTicket)

		admin.PUT("/wisata/:id", adminC.UpdateWisata)
		admin.DELETE("/wisata/:id", adminC.DeleteWisata)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
