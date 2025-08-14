package services

import (
	"context"
	"errors"
	"fmt"
	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/domain"
	"lawnconnect-api/internal/infrastructure/database/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BookingService defines the service interface for bookings.
type BookingService interface {
	CreateBooking(ctx context.Context, customerID primitive.ObjectID, date, bookingTime, address, description string) (*domain.Booking, error)
	GetBookingByID(ctx context.Context, bookingID primitive.ObjectID) (*domain.Booking, error)
	ListBookings(ctx context.Context, userID primitive.ObjectID) ([]*domain.Booking, error)
	ListPendingBookings(ctx context.Context) ([]*domain.Booking, error)
	AcceptBooking(ctx context.Context, bookingID, mowerID primitive.ObjectID) error
	RejectBooking(ctx context.Context, bookingID, mowerID primitive.ObjectID) error
	CompleteBooking(ctx context.Context, bookingID primitive.ObjectID, price float64) error
	CancelBooking(ctx context.Context, bookingID, customerID primitive.ObjectID) error
}

type bookingService struct {
	bookingRepo repositories.BookingRepository
}

// NewBookingService creates a new BookingService.
func NewBookingService(bookingRepo repositories.BookingRepository) BookingService {
	return &bookingService{bookingRepo: bookingRepo}
}

// CreateBooking creates a new booking.
func (s *bookingService) CreateBooking(ctx context.Context, customerID primitive.ObjectID, date, bookingTime, address, description string) (*domain.Booking, error) {
	// Simple validation
	if date == "" || bookingTime == "" || address == "" {
		return nil, errors.New("date, time, and address are required")
	}

	booking := &domain.Booking{
		ID:          primitive.NewObjectID(),
		CustomerID:  customerID,
		Date:        date,
		Time:        bookingTime,
		Address:     address,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.bookingRepo.CreateBooking(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("service failed to create booking: %w", err)
	}

	return booking, nil
}

// GetBookingByID retrieves a single booking by ID.
func (s *bookingService) GetBookingByID(ctx context.Context, bookingID primitive.ObjectID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		if _, ok := err.(apperror.NotFound); ok {
			return nil, err
		}
		return nil, fmt.Errorf("service failed to get booking: %w", err)
	}
	return booking, nil
}

// ListBookings retrieves all bookings for a user.
func (s *bookingService) ListBookings(ctx context.Context, userID primitive.ObjectID) ([]*domain.Booking, error) {
	bookings, err := s.bookingRepo.FindBookingsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service failed to list bookings: %w", err)
	}
	return bookings, nil
}

// ListPendingBookings retrieves all bookings with a pending status.
func (s *bookingService) ListPendingBookings(ctx context.Context) ([]*domain.Booking, error) {
	bookings, err := s.bookingRepo.FindPendingBookings(ctx)
	if err != nil {
		return nil, fmt.Errorf("service failed to list pending bookings: %w", err)
	}
	return bookings, nil
}

// AcceptBooking handles a mower accepting a booking.
func (s *bookingService) AcceptBooking(ctx context.Context, bookingID, mowerID primitive.ObjectID) error {
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != "pending" {
		return errors.New("booking is not pending and cannot be accepted")
	}
	if booking.MowerID != primitive.NilObjectID && booking.MowerID != mowerID {
		return apperror.CustomError{Message: "This booking has already been accepted by another mower"}
	}

	update := bson.M{
		"$set": bson.M{
			"mowerId":   mowerID,
			"status":    "accepted",
			"updatedAt": time.Now(),
		},
	}
	err = s.bookingRepo.UpdateBooking(ctx, bookingID, update)
	if err != nil {
		return fmt.Errorf("service failed to accept booking: %w", err)
	}
	return nil
}

// RejectBooking handles a mower rejecting a booking.
func (s *bookingService) RejectBooking(ctx context.Context, bookingID, mowerID primitive.ObjectID) error {
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != "pending" {
		return errors.New("booking is not pending and cannot be rejected")
	}

	// This check ensures a mower can only reject jobs that haven't been accepted by another mower.
	if booking.MowerID != primitive.NilObjectID && booking.MowerID != mowerID {
		return apperror.CustomError{Message: "This booking has already been accepted by another mower"}
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "rejected",
			"updatedAt": time.Now(),
		},
	}
	err = s.bookingRepo.UpdateBooking(ctx, bookingID, update)
	if err != nil {
		return fmt.Errorf("service failed to reject booking: %w", err)
	}
	return nil
}

// CompleteBooking handles a mower completing a booking.
func (s *bookingService) CompleteBooking(ctx context.Context, bookingID primitive.ObjectID, price float64) error {
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != "accepted" {
		return apperror.CustomError{Message: "Booking is not accepted and cannot be completed"}
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "completed",
			"price":     price,
			"updatedAt": time.Now(),
		},
	}
	err = s.bookingRepo.UpdateBooking(ctx, bookingID, update)
	if err != nil {
		return fmt.Errorf("service failed to complete booking: %w", err)
	}
	return nil
}

// CancelBooking handles a customer cancelling a booking.
func (s *bookingService) CancelBooking(ctx context.Context, bookingID, customerID primitive.ObjectID) error {
	booking, err := s.bookingRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != "pending" && booking.Status != "accepted" {
		return apperror.CustomError{Message: "Booking cannot be cancelled in its current state"}
	}

	// Ensure the user cancelling is the original customer
	if booking.CustomerID != customerID {
		return apperror.CustomError{Message: "Unauthorized to cancel this booking"}
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "cancelled",
			"updatedAt": time.Now(),
		},
	}
	err = s.bookingRepo.UpdateBooking(ctx, bookingID, update)
	if err != nil {
		return fmt.Errorf("service failed to cancel booking: %w", err)
	}
	return nil
}
