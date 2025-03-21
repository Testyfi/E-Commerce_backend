package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type contextKey string

const ContextUser contextKey = "user"

type User struct {
	ID             primitive.ObjectID           `bson:"_id"`
	First_name     *string                      `json:"first_name" validate:"max=100,required"`
	Last_name      *string                      `json:"last_name" validate:"max=100,required"`
	Password       *string                      `json:"password" validate:"required,min=6"`
	Phone          *string                      `json:"phone" validate:"required,len=10,numeric"`
	Token          *string                      `json:"token"`
	Refresh_token  *string                      `json:"refresh_token"`
	Created_at     time.Time                    `json:"created_at"`
	Updated_at     time.Time                    `json:"updated_at"`
	User_id        string                       `json:"user_id"`
	QuestionPapers map[string]map[string]string `json:"questionPapers"`
	Profile        string                       `json:"profile"`
	ReferralCode   string                       `json:"referral_code"`
	Wallet         int                          `json:"wallet"`
	Purchased      bool                         `json:"purchased"`
	PurchaseDate   time.Time                    `json:"purchasedate"`
	PurchasePlan   int                          `json:"purchaseplane"`
	RankBoosterTest int                         `json:"rankboostertest"`
	CreateYourTest  int                         `json:"createyourtest"`
	Verified       bool                         `json:"verified"`
	ResetCode      string                       `json:"reset_code"`
	SecretCode     string                       `json:"secret_code"`
	Otp            string                       `json:"otp"`
}

type OTP struct {
	Email      *string   `json:"email" validate:"email,required"`
	Phone      *string   `json:"phone" validate:"required,len=10,numeric"`
	First_name *string   `json:"first_name" validate:"max=100,required"`
	SecretCode string    `json:"secret_code"`
	Otp        string    `json:"otp"`
	CreatedAt  time.Time `json:"createdat"`
}
