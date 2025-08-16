package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect deschide conexiunea la Mongo şi rulează un ping.
func Connect(uri string) (*mongo.Client, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        return nil, err
    }
    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }
    return client, nil
}

// UserCollection returnează handle-ul pe colecţia "users".
func UserCollection(client *mongo.Client) *mongo.Collection {
    return client.Database("API-GO").Collection("users")
}

// BookCollection returns a handle to the "books" collection.
func BookCollection(client *mongo.Client) *mongo.Collection {
    return client.Database("API-GO").Collection("books")
}

// CreateIndexes creează indecși unici pentru email și phone
func CreateIndexes(client *mongo.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Users indexes
    coll := UserCollection(client)
    
    // Index unic pentru email
    emailIndex := mongo.IndexModel{
        Keys:    bson.M{"email": 1},
        Options: options.Index().SetUnique(true),
    }
    
    // Index unic pentru phone (doar dacă nu e null/empty)
    phoneIndex := mongo.IndexModel{
        Keys: bson.M{"phone": 1},
        Options: options.Index().SetUnique(true).SetSparse(true), // sparse pentru câmpuri opționale
    }

    if _, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{emailIndex, phoneIndex}); err != nil {
        return err
    }

    // Books indexes (pentru căutări/filtrări frecvente)
    bcoll := BookCollection(client)
    bookIndexes := []mongo.IndexModel{
        {Keys: bson.M{"title": 1}},
        {Keys: bson.M{"author": 1}},
        //{Keys: bson.M{"genre": 1}},
        //{Keys: bson.M{"yearPublished": 1}},
    }
    _, err := bcoll.Indexes().CreateMany(ctx, bookIndexes)
    return err
}