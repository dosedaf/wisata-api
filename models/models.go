package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Email        string    `gorm:"uniqueIndex" json:"email"`
	Password     string    `json:"-"`
	Role         string    `gorm:"default:'USER'" json:"role"`
	PhoneNumber  string    `json:"phoneNumber"`
	ProfileImage *string   `json:"profileImage"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Wisata struct {
	ID                   uint            `gorm:"primaryKey" json:"id"`
	Name                 string          `json:"name"`
	Slug                 string          `gorm:"uniqueIndex" json:"slug"`
	Location             string          `json:"location"`
	Description          string          `json:"description"`
	TicketPrice          float64         `json:"ticketPrice"`
	Capacity             int             `json:"capacity"`
	Rating               float64         `gorm:"default:0" json:"rating"`
	MapsUrl              string          `json:"mapsUrl"`
	IsDisabilityFriendly bool            `json:"isDisabilityFriendly"`
	IsKidsFree           bool            `json:"isKidsFree"`
	ImageUrl             string          `json:"imageUrl"`
	Schedules            []Schedule      `gorm:"foreignKey:WisataID" json:"-"`
	Galleries            []WisataGallery `gorm:"foreignKey:WisataID" json:"galleries,omitempty"`
	Tags                 []Tag           `gorm:"many2many:wisata_tags;" json:"tags,omitempty"`
	Reviews              []Review        `gorm:"foreignKey:WisataID" json:"reviews,omitempty"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"-"`
	Name string `gorm:"uniqueIndex" json:"name"`
}

type WisataGallery struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	WisataID uint   `json:"-"`
	ImageUrl string `json:"imageUrl"`
}

type Schedule struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	WisataID       uint      `json:"wisataId"`
	VisitDate      string    `json:"visitDate"` // Gunakan string YYYY-MM-DD agar sederhana
	Quota          int       `json:"quota"`
	RemainingQuota int       `json:"remainingQuota"`
}

type Booking struct {
	ID          uint      `gorm:"primaryKey" json:"bookingId"`
	UserID      uint      `json:"userId"`
	WisataID    uint      `json:"wisataId"`
	Wisata      Wisata    `gorm:"foreignKey:WisataID" json:"wisata"`
	ScheduleID  uint      `json:"scheduleId"`
	Schedule    Schedule  `gorm:"foreignKey:ScheduleID" json:"schedule"`
	BookingCode string    `gorm:"unique" json:"bookingCode"`
	TotalTicket int       `json:"totalTicket"`
	TotalPrice  float64   `json:"totalPrice"`
	Status      string    `gorm:"default:'PENDING'" json:"status"` // PENDING, ACTIVE, COMPLETED
	QRCode      string    `json:"qrCode"`
	HasReviewed bool      `gorm:"default:false" json:"hasReviewed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Review struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WisataID  uint      `json:"-"`
	UserID    uint      `json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	BookingID uint      `json:"-"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}
