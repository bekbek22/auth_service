package repository

import (
	"context"
	"time"

	"github.com/bekbek22/auth_service/internal/model"

	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	user.CreatedAt = time.Now().Unix()
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, map[string]interface{}{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
