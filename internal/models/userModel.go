package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
    ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name     string             `bson:"name" json:"name"`
    Email    string             `bson:"email" json:"email"`
    Password string             `bson:"password,omitempty" json:"-"`
    Phone    string             `bson:"phone,omitempty" json:"phone"`
}

// DTO pentru Create/Update
type UserInput struct {
    Name            string `json:"name"`
    Email           string `json:"email"`
    Password        string `json:"password"`
    PasswordConfirm string `json:"passwordConfirm"`
    Phone           string `json:"phone"`
}