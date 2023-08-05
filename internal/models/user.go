package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID           `bson:"_id"`
	First_name     *string                      `json:"first_name" validate:"max=100,required"`
	Last_name      *string                      `json:"last_name" validate:"max=100,required"`
	Password       *string                      `json:"password" validate:"required,min=6"`
	Email          *string                      `json:"email" validate:"email,required"`
	Phone          *string                      `json:"phone" validate:"required,len=10,numeric"`
	Token          *string                      `json:"token"`
	Refresh_token  *string                      `json:"refresh_token"`
	Created_at     time.Time                    `json:"created_at"`
	Updated_at     time.Time                    `json:"updated_at"`
	User_id        string                       `json:"user_id"`
	QuestionPapers map[string]map[string]string `json:"questionPapers"`
}
