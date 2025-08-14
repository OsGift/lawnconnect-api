package repositories

import (
	"context"
	"fmt"
	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// BookingRepository defines the repository interface for bookings.
type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *domain.Booking) error
	FindBookingByID(ctx context.Context, bookingID primitive.ObjectID) (*domain.Booking, error)
	FindBookingsByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Booking, error)
	FindPendingBookings(ctx context.Context) ([]*domain.Booking, error)
	UpdateBooking(ctx context.Context, bookingID primitive.ObjectID, update bson.M) error
}

type bookingRepository struct {
	collection *mongo.Collection
}

// NewBookingRepository creates a new BookingRepository.
func NewBookingRepository(db *mongo.Database) BookingRepository {
	return &bookingRepository{collection: db.Collection("bookings")}
}

// CreateBooking inserts a new booking document into the database.
func (r *bookingRepository) CreateBooking(ctx context.Context, booking *domain.Booking) error {
	_, err := r.collection.InsertOne(ctx, booking)
	if err != nil {
		return fmt.Errorf("failed to insert booking: %w", err)
	}
	return nil
}

// FindBookingByID retrieves a single booking by its unique ID.
func (r *bookingRepository) FindBookingByID(ctx context.Context, bookingID primitive.ObjectID) (*domain.Booking, error) {
	var booking domain.Booking
	filter := bson.M{"_id": bookingID}
	err := r.collection.FindOne(ctx, filter).Decode(&booking)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperror.NotFound{Resource: "Booking"}
		}
		return nil, fmt.Errorf("failed to find booking: %w", err)
	}
	return &booking, nil
}

// FindBookingsByUserID retrieves all bookings for a given user, whether they are the customer or the mower.
func (r *bookingRepository) FindBookingsByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Booking, error) {
	var bookings []*domain.Booking
	filter := bson.M{
		"$or": []bson.M{
			{"customerId": userID},
			{"mowerId": userID},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find bookings: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, fmt.Errorf("failed to decode bookings: %w", err)
	}
	return bookings, nil
}

// FindPendingBookings retrieves all bookings with a "pending" status.
func (r *bookingRepository) FindPendingBookings(ctx context.Context) ([]*domain.Booking, error) {
	var bookings []*domain.Booking
	filter := bson.M{"status": "pending"}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending bookings: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, fmt.Errorf("failed to decode pending bookings: %w", err)
	}
	return bookings, nil
}


// UpdateBooking updates a booking document by its ID.
func (r *bookingRepository) UpdateBooking(ctx context.Context, bookingID primitive.ObjectID, update bson.M) error {
	filter := bson.M{"_id": bookingID}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}
	return nil
}
