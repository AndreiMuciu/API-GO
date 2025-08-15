package repository

import (
	"API-GO/internal/models"
	"API-GO/internal/utils"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BookRepository follows the CRUDRepository contract for Book.
type BookRepository interface {
    CRUDRepository[models.Book, primitive.ObjectID]
	// Extended list supporting filtering/sorting/pagination
	ListWithQuery(ctx context.Context, q utils.ListQuery) ([]models.Book, int64, error)
}
