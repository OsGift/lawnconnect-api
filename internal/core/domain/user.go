package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system (customer, mower, admin, super_admin).
type User struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                string             `bson:"name" json:"name" validate:"required"`
	Email               string             `bson:"email" json:"email" validate:"required,email"`
	Password            string             `bson:"password" json:"-" validate:"required,min=6"` // "-" omits from JSON output
	Role                string             `bson:"role" json:"role" validate:"required,oneof=customer mower admin super_admin"`
	IsVerified          bool               `bson:"isVerified" json:"isVerified"`
	DefaultPassword     bool               `bson:"defaultPassword,omitempty" json:"defaultPassword,omitempty"` // For admins
	ImageUrl            string             `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
	PhoneNumber         string             `bson:"phoneNumber,omitempty" json:"phoneNumber,omitempty"`
	BusinessAddress     string             `bson:"businessAddress,omitempty" json:"businessAddress,omitempty"`
	ContactPerson       string             `bson:"contactPerson,omitempty" json:"contactPerson,omitempty"`
	ContactPersonEmail  string             `bson:"contactPersonEmail,omitempty" json:"contactPersonEmail,omitempty"`
	ContactPersonPhone  string             `bson:"contactPersonPhone,omitempty" json:"contactPersonPhone,omitempty"`
	BusinessPhoneNumber string             `bson:"businessPhoneNumber,omitempty" json:"businessPhoneNumber,omitempty"`
	BusinessEmail       string             `bson:"businessEmail,omitempty" json:"businessEmail,omitempty"`
	IsApproved          bool               `bson:"isApproved" json:"isApproved"`   // For 'mower' role
	IsAvailable         bool               `bson:"isAvailable" json:"isAvailable"` // For 'mower' role
	Services            []string           `bson:"services,omitempty" json:"services,omitempty"`
	Availability        []UserAvailability `bson:"availability,omitempty" json:"availability,omitempty"`
	HourlyRate          float64            `bson:"hourlyRate" json:"hourlyRate"` // For 'mower' role
	Ratings             []UserRating       `bson:"ratings,omitempty" json:"ratings,omitempty"`
	WalletBalance       float64            `bson:"walletBalance" json:"walletBalance"` // For 'mower' role
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt" json:"updatedAt"`
	ResetTokenExpiresAt time.Time          `bson:"resetTokenExpiresAt,omitempty" json:"resetTokenExpiresAt,omitempty"` // For password reset
}

// UserAvailability represents a time slot a mower is available.
type UserAvailability struct {
	Day      string `bson:"day" json:"day" validate:"required"`
	FromTime string `bson:"fromTime" json:"fromTime" validate:"required"` // "HH:MM" format
	ToTime   string `bson:"toTime" json:"toTime" validate:"required"`     // "HH:MM" format
}

// UserRating represents a rating given to a mower by a customer.
type UserRating struct {
	CustomerID primitive.ObjectID `bson:"customerId" json:"customerId" validate:"required"`
	Rating     int                `bson:"rating" json:"rating" validate:"min=1,max=5"`
	Comment    string             `bson:"comment,omitempty" json:"comment,omitempty"`
	BookingID  primitive.ObjectID `bson:"bookingId" json:"bookingId" validate:"required"`
	IsRating   bool               `bson:"isRating" json:"isRating"` // True if this comment is part of a direct rating
}
