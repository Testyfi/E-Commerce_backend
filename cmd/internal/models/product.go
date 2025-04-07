package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)


type Product struct {
	ID             primitive.ObjectID `bson:"_id"`
	Product       string             `json:"productName"`
	Images         []string           `json:"images"`
	Created_at     time.Time          `bson:"created_at"`
	Product_Tags   []string           `json:"productTags"`
	P_id           string             `json:"pid"`
	Sml            int                `json:"sml"`
	Md            int                `json:"md"`
	Lar            int                `json:"lar"`
	Xl            int                `json:"xl"`
    Mrp            int                `json:"mrp"`
	Price            int                `json:"price"`
}


