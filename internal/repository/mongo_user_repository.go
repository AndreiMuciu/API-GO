package repository

import (
	"context"

	"API-GO/internal/database"
	"API-GO/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUserRepository struct {
    client *mongo.Client
}

func NewMongoUserRepository(client *mongo.Client) *MongoUserRepository {
    return &MongoUserRepository{client: client}
}

func (r *MongoUserRepository) collection() *mongo.Collection {
    return database.UserCollection(r.client)
}

func (r *MongoUserRepository) Create(ctx context.Context, user *models.User) error {
    _, err := r.collection().InsertOne(ctx, user)
    return err
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
    var user models.User
    err := r.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.collection().FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *MongoUserRepository) List(ctx context.Context) ([]models.User, error) {
    cur, err := r.collection().Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var users []models.User
    if err := cur.All(ctx, &users); err != nil {
        return nil, err
    }
    return users, nil
}

func (r *MongoUserRepository) EmailExists(ctx context.Context, email string, excludeID ...primitive.ObjectID) (bool, error) {
    filter := bson.M{"email": email}
    if len(excludeID) > 0 && excludeID[0] != primitive.NilObjectID {
        filter["_id"] = bson.M{"$ne": excludeID[0]}
    }
    count, err := r.collection().CountDocuments(ctx, filter)
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func (r *MongoUserRepository) PhoneExists(ctx context.Context, phone string, excludeID ...primitive.ObjectID) (bool, error) {
    if phone == "" {
        return false, nil
    }
    filter := bson.M{"phone": phone}
    if len(excludeID) > 0 && excludeID[0] != primitive.NilObjectID {
        filter["_id"] = bson.M{"$ne": excludeID[0]}
    }
    count, err := r.collection().CountDocuments(ctx, filter)
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func (r *MongoUserRepository) UpdateFields(ctx context.Context, id primitive.ObjectID, fields map[string]interface{}) (bool, error) {
    res, err := r.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})
    if err != nil {
        return false, err
    }
    return res.MatchedCount > 0, nil
}

func (r *MongoUserRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
    res, err := r.collection().DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return false, err
    }
    return res.DeletedCount > 0, nil
}
