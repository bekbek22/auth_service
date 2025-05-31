package repository

import (
	"context"
	"time"

	"github.com/bekbek22/auth_service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IUserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	UpdateUserByID(ctx context.Context, id primitive.ObjectID, update interface{}) error
	SoftDeleteUserByID(ctx context.Context, id primitive.ObjectID) error
}
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
	err := r.collection.FindOne(ctx, bson.M{
		"email":      email,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindUsers(ctx context.Context, name, email string, page, limit int32) ([]model.User, int32, error) {
	filter := bson.M{
		"is_deleted": bson.M{"$ne": true},
	}
	if name != "" {
		filter["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	if email != "" {
		filter["email"] = bson.M{"$regex": email, "$options": "i"}
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	count, _ := r.collection.CountDocuments(ctx, filter)

	return users, int32(count), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateUserByID(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	return err
}

func (r *UserRepository) SoftDeleteUserByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
	return err
}
