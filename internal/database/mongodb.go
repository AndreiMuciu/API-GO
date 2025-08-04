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

// CheckEmailExists verifică dacă emailul există deja în baza de date
func CheckEmailExists(client *mongo.Client, email string, excludeID ...string) (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    filter := bson.M{"email": email}
    
    // Dacă actualizăm un user existent, excludem ID-ul său din căutare
    if len(excludeID) > 0 && excludeID[0] != "" {
        filter["_id"] = bson.M{"$ne": excludeID[0]}
    }

    coll := UserCollection(client)
    count, err := coll.CountDocuments(ctx, filter)
    if err != nil {
        return false, err
    }
    
    return count > 0, nil
}

// CheckPhoneExists verifică dacă numărul de telefon există deja în baza de date
func CheckPhoneExists(client *mongo.Client, phone string, excludeID ...string) (bool, error) {
    if phone == "" {
        return false, nil // telefonul nu e obligatoriu
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    filter := bson.M{"phone": phone}
    
    // Dacă actualizăm un user existent, excludem ID-ul său din căutare
    if len(excludeID) > 0 && excludeID[0] != "" {
        filter["_id"] = bson.M{"$ne": excludeID[0]}
    }

    coll := UserCollection(client)
    count, err := coll.CountDocuments(ctx, filter)
    if err != nil {
        return false, err
    }
    
    return count > 0, nil
}

// CreateIndexes creează indecși unici pentru email și phone
func CreateIndexes(client *mongo.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

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

    _, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{emailIndex, phoneIndex})
    return err
}