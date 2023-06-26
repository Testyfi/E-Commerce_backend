package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Option struct {
	Text  string `json:"text"`
	Image string `json:"image"`
}
type Question struct {
	ID            primitive.ObjectID `bson:"_id"`
	Question      string             `json:"questionText"`
	Images        []string           `json:"images"`
	Type          string             `json:"questionType"`
	Options       []Option           `json:"options"`
	CorrectAnswer string             `json:"correctAnswer"`
	Created_at    time.Time          `bson:"created_at"`
	Subject_Tags  []string           `json:"subjectTags"`
	Q_id          string             `json:"qid"`
}
