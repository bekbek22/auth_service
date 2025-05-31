package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Role      string             `bson:"role"`
	IsDeleted bool               `bson:"is_deleted"` //สำหรับ soft delete
	CreatedAt int64              `bson:"created_at"`
}

type BlacklistedToken struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	JTI           string             `bson:"jti"`            // JWT ID from the token
	ExpiresAt     time.Time          `bson:"expires_at"`     // Original expiration time of the token
	BlacklistedAt time.Time          `bson:"blacklisted_at"` // When it was blacklisted
}
