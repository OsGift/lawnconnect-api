package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SystemMetrics is a document to store global metrics like total platform revenue.
type SystemMetrics struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string             `bson:"name" json:"name"`   // e.g., "system_revenue"
	Value float64            `bson:"value" json:"value"` // The actual revenue amount
}