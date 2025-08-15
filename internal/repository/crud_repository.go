package repository

import "context"

// CRUDRepository defines a generic interface for simple CRUD.
// T is the domain model and ID is its identifier type.
// In Go we don't have type parameters in interfaces across packages for runtime use here,
// so we'll define a non-generic contract for common operations and implement per entity.
// This file serves as a guideline for standard method names.
type CRUDRepository[T any, ID any] interface {
    Create(ctx context.Context, entity *T) error
    GetByID(ctx context.Context, id ID) (*T, error)
    List(ctx context.Context) ([]T, error)
    UpdateFields(ctx context.Context, id ID, fields map[string]interface{}) (bool, error)
    DeleteByID(ctx context.Context, id ID) (bool, error)
}
