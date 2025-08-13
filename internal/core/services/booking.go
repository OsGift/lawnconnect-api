package services

import (
	"context"
	"fmt"
	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/domain"
	"lawnconnect-api/internal/infrastructure/database/repositories"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

// BookingService defines the business logic for bookings.
type BookingService interface {
	CreateBooking(ctx context.Context, customerID primitive.ObjectID, date, timeStr, address, description string) (*domain.Booking, error)
	GetBookingByID(ctx context.Context, bookingID primitive.ObjectID) (*domain.Booking, error)
	AcceptBooking(ctx context.Context, bookingID, mowerID primitive.ObjectID) error
	CompleteBooking(ctx context.Context, bookingID primitive.ObjectID, price float64) error
	ListBookings(ctx context.Context, userID primitive.ObjectID) ([]*domain.Booking, error)
	CancelBooking(ctx context.Context, bookingID, customerID primitive.ObjectID) error
}

type bookingService struct {
	bookingRepo repositories.BookingRepository
}

// NewBookingService creates a new BookingService instance.
func NewBookingService(bookingRepo repositories.BookingRepository) BookingService {
	return &bookingService{bookingRepo: bookingRepo}
}

// CreateBooking creates a new booking with a 'pending' status.
func (s *bookingService) CreateBooking(ctx context.Context, customerID primitive.ObjectID, date, timeStr, address, description string) (*domain.Booking, error) {
	booking := &domain.Booking{
		ID:            primitive.NewObjectID(),
		CustomerID:    customerID,
		Date:          date,
		Time:          timeStr,
		Address:       address,
		Description:   description,
		Status:        "pending",
		Price:         0.0, // Price will be set by the mower
		BillingStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.bookingRepo.CreateBooking(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to create booking in repository: %w", err)
	}

	return booking, nil
}

// GetBookingByID retrieves a single booking by its ID.
func (s *bookingService) GetBookingByID(ctx context.Context, bookingID primitive.ObjectID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve booking: %w", err)
	}
	return booking, nil
}

// AcceptBooking updates a booking's status to 'accepted' and assigns a mower.
func (s *bookingService) AcceptBooking(ctx context.Context, bookingID, mowerID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"status":       "accepted",
			"mowerId":      mowerID,
			"acceptedTime": time.Now(),
			"updatedAt":    time.Now(),
		},
	}
	return s.bookingRepo.UpdateBooking(ctx, bookingID, update)
}

// CompleteBooking handles the finalization of a booking, including payment simulation and status updates.
func (s *bookingService) CompleteBooking(ctx context.Context, bookingID primitive.ObjectID, price float64) error {
	// First, retrieve the booking to ensure it exists and is in an 'accepted' state
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to find booking for completion: %w", err)
	}

	if booking.Status != "accepted" {
		return apperror.CustomError{Message: "booking cannot be completed as it is not in an accepted status"}
	}

	// --- Payment Simulation ---
	// In a real-world scenario, you would integrate with a payment gateway here.
	// For this simulation, we'll assume the payment is successful.
	log.Printf("Simulating payment for booking %s with amount $%.2f", bookingID.Hex(), price)
	time.Sleep(2 * time.Second) // Simulate a network call to a payment gateway

	// --- Update Booking Status and Invoice ---
	update := bson.M{
		"$set": bson.M{
			"status":        "completed",
			"price":         price,
			"billingStatus": "paid",
			"completedAt":   time.Now(),
			"updatedAt":     time.Now(),
		},
	}

	return s.bookingRepo.UpdateBooking(ctx, bookingID, update)
}

// ListBookings retrieves all bookings for a given user, whether they are the customer or the mower.
func (s *bookingService) ListBookings(ctx context.Context, userID primitive.ObjectID) ([]*domain.Booking, error) {
	bookings, err := s.bookingRepo.FindBookingsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings in repository: %w", err)
	}
	return bookings, nil
}

// CancelBooking cancels a booking by its ID.
func (s *bookingService) CancelBooking(ctx context.Context, bookingID, customerID primitive.ObjectID) error {
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	if booking.CustomerID != customerID {
		return apperror.CustomError{Message: "Unauthorized: only the customer who created the booking can cancel it"}
	}

	if booking.Status == "completed" || booking.Status == "cancelled" {
		return apperror.CustomError{Message: "Booking is already completed or cancelled and cannot be cancelled again"}
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "cancelled",
			"updatedAt": time.Now(),
		},
	}
	return s.bookingRepo.UpdateBooking(ctx, bookingID, update)
}
