package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type ITokenRepository interface {
	BlacklistToken(ctx context.Context, token string, exp int64) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}
type TokenRepository struct {
	collection *mongo.Collection
}

func NewTokenRepository(db *mongo.Database) *TokenRepository {
	return &TokenRepository{
		collection: db.Collection("blacklisted_tokens"),
	}
}

func (r *TokenRepository) BlacklistToken(ctx context.Context, token string, exp int64) error {
	doc := map[string]interface{}{
		"token": token,
		"exp":   exp,
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *TokenRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, map[string]interface{}{
		"token": token,
	})
	return count > 0, err
}
