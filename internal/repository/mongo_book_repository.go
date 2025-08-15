package repository

import (
    "context"
    "sync"

    "API-GO/internal/models"
    "API-GO/internal/utils"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoBookRepository struct {
    client *mongo.Client
}

func NewMongoBookRepository(client *mongo.Client) *MongoBookRepository {
    return &MongoBookRepository{client: client}
}

func (r *MongoBookRepository) collection() *mongo.Collection {
    return r.client.Database("API-GO").Collection("books")
}

func (r *MongoBookRepository) Create(ctx context.Context, b *models.Book) error {
    _, err := r.collection().InsertOne(ctx, b)
    return err
}

func (r *MongoBookRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Book, error) {
    var b models.Book
    err := r.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&b)
    if err != nil {
        return nil, err
    }
    return &b, nil
}

func (r *MongoBookRepository) List(ctx context.Context) ([]models.Book, error) {
    cur, err := r.collection().Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var out []models.Book
    if err := cur.All(ctx, &out); err != nil {
        return nil, err
    }
    return out, nil
}

func (r *MongoBookRepository) ListWithQuery(ctx context.Context, q utils.ListQuery) ([]models.Book, int64, error) {
    opts := options.Find()
    if len(q.Sort) > 0 {
        opts.SetSort(q.Sort)
    }
    if q.Limit > 0 {
        opts.SetLimit(q.Limit)
        opts.SetSkip(q.Skip)
    }
    filter := q.Filter
    if filter == nil {
        filter = bson.M{}
    }

    var wg sync.WaitGroup
    wg.Add(2)
    var (
        items []models.Book
        total int64
        findErr error
        countErr error
    )
    // Fetch items
    go func() {
        defer wg.Done()
        cur, err := r.collection().Find(ctx, filter, opts)
        if err != nil { findErr = err; return }
        defer cur.Close(ctx)
        var out []models.Book
        if err := cur.All(ctx, &out); err != nil { findErr = err; return }
        items = out
    }()
    // Count total
    go func() {
        defer wg.Done()
        cnt, err := r.collection().CountDocuments(ctx, filter)
        if err != nil { countErr = err; return }
        total = cnt
    }()
    wg.Wait()
    if findErr != nil { return nil, 0, findErr }
    if countErr != nil { return nil, 0, countErr }
    return items, total, nil
}

func (r *MongoBookRepository) UpdateFields(ctx context.Context, id primitive.ObjectID, fields map[string]interface{}) (bool, error) {
    res, err := r.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})
    if err != nil {
        return false, err
    }
    return res.MatchedCount > 0, nil
}

func (r *MongoBookRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
    res, err := r.collection().DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return false, err
    }
    return res.DeletedCount > 0, nil
}
