package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Booking represents a lawn mowing service booking.
type Booking struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID           primitive.ObjectID `bson:"customerId" json:"customerId" validate:"required"`
	MowerID              primitive.ObjectID `bson:"mowerId,omitempty" json:"mowerId,omitempty"` // Omitted if unassigned
	Date                 string             `bson:"date" json:"date" validate:"required"`       // YYYY-MM-DD
	Time                 string             `bson:"time" json:"time" validate:"required"`       // HH:MM
	Address              string             `bson:"address" json:"address" validate:"required"`
	Description          string             `bson:"description,omitempty" json:"description,omitempty"`
	Status               string             `bson:"status" json:"status" validate:"required,oneof=pending accepted ongoing completed cancelled rejected"`
	Price                float64            `bson:"price" json:"price"`
	BillingStatus        string             `bson:"billingStatus" json:"billingStatus" validate:"required,oneof=pending billed paid"`
	Rating               int                `bson:"rating,omitempty" json:"rating,omitempty"` // Overall rating for the booking
	Comments             []BookingComment   `bson:"comments,omitempty" json:"comments,omitempty"` // New array for all comments
	AcceptedTime         *time.Time         `bson:"acceptedTime,omitempty" json:"acceptedTime,omitempty"`
	OngoingTime          *time.Time         `bson:"ongoingTime,omitempty" json:"ongoingTime,omitempty"`
	CompletedTime        *time.Time         `bson:"completedTime,omitempty" json:"completedTime,omitempty"`
	ProofOfCompletionURL string             `bson:"proofOfCompletionUrl,omitempty" json:"proofOfCompletionUrl,omitempty"`
	CompletionComment    string             `bson:"completionComment,omitempty" json:"completionComment,omitempty"`
	RejectionReason      string             `bson:"rejectionReason,omitempty" json:"rejectionReason,omitempty"`
	PaymentReminderSent  bool               `bson:"paymentReminderSent" json:"paymentReminderSent"` // For invoice simulation
	CreatedAt            time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// BookingComment represents a single comment on a booking.
type BookingComment struct {
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	UserName    string             `bson:"userName" json:"userName"`
	CommentText string             `bson:"commentText" json:"commentText"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
	IsRating    bool               `bson:"isRating,omitempty" json:"isRating,omitempty"` // Indicates if this comment is also a rating comment
	Rating      int                `bson:"rating,omitempty" json:"rating,omitempty"`     // Rating if IsRating is true
}