package repository

import (
	"context"

	"API-GO/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the contract for user data access (SOLID: DIP)
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    EmailExists(ctx context.Context, email string, excludeID ...primitive.ObjectID) (bool, error)
    PhoneExists(ctx context.Context, phone string, excludeID ...primitive.ObjectID) (bool, error)
    List(ctx context.Context) ([]models.User, error)
    UpdateFields(ctx context.Context, id primitive.ObjectID, fields map[string]interface{}) (bool, error)
    DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error)
}
