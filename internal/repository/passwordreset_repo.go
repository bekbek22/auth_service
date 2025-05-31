package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type IPasswordResetRepository interface {
	SaveToken(ctx context.Context, email, token string, exp int64) error
	GetEmailByToken(ctx context.Context, token string) (string, error)
	DeleteToken(ctx context.Context, token string) error
}

type PasswordResetRepository struct {
	collection *mongo.Collection
}

func NewPasswordResetRepository(db *mongo.Database) *PasswordResetRepository {
	return &PasswordResetRepository{
		collection: db.Collection("password_resets"),
	}
}

func (r *PasswordResetRepository) SaveToken(ctx context.Context, email, token string, exp int64) error {
	doc := map[string]interface{}{
		"email": email,
		"token": token,
		"exp":   exp,
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *PasswordResetRepository) GetEmailByToken(ctx context.Context, token string) (string, error) {
	now := time.Now().Unix()
	var result struct {
		Email string `bson:"email"`
		Exp   int64  `bson:"exp"`
	}
	err := r.collection.FindOne(ctx, map[string]interface{}{
		"token": token,
		"exp":   map[string]interface{}{"$gt": now},
	}).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.Email, nil
}

func (r *PasswordResetRepository) DeleteToken(ctx context.Context, token string) error {
	_, err := r.collection.DeleteOne(ctx, map[string]interface{}{
		"token": token,
	})
	return err
}
