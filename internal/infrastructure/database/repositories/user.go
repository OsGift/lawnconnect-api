package repositories

import (
	"context"

	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepository defines the interface for interacting with user data.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	FindUserByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	FindUserByResetToken(ctx context.Context, token string) (*domain.User, error)
	UpdateUser(ctx context.Context, id primitive.ObjectID, update primitive.M) error
}

type userRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{collection: db.Collection("users")}
}

// CreateUser saves a new user to the database.
func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindUserByEmail retrieves a user by their email address.
func (r *userRepository) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	filter := primitive.M{"email": email}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperror.NotFound{Resource: "User"}
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByID retrieves a user by their ID.
func (r *userRepository) FindUserByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	filter := primitive.M{"_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperror.NotFound{Resource: "User"}
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByResetToken retrieves a user by their password reset token.
func (r *userRepository) FindUserByResetToken(ctx context.Context, token string) (*domain.User, error) {
	var user domain.User
	filter := primitive.M{"resetToken": token}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperror.NotFound{Resource: "User"}
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user's document with the provided BSON update.
func (r *userRepository) UpdateUser(ctx context.Context, id primitive.ObjectID, update primitive.M) error {
	filter := primitive.M{"_id": id}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
