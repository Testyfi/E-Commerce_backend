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
	ID             primitive.ObjectID `bson:"_id"`
	Question       string             `json:"questionText"`
	Images         []string           `json:"images"`
	Type           string             `json:"questionType"`
	Options        []Option           `json:"options"`
	CorrectAnswer  string             `json:"correctAnswer"`
	CorrectAnswers []string           `json:"correctAnswers"`
	Created_at     time.Time          `bson:"created_at"`
	Subject_Tags   []string           `json:"subjectTags"`
	Q_id           string             `json:"qid"`
	UsedBy         []string           `json:"usedBy"`
	Solution       string             `json:"solution"`
	List1          []string           `json:"list1"`
	List2          []string           `json:"list2"`
}

type QPaper struct {
	Name string        `json:"name"`
	Difficulty string  `json:difficulty`
	Duration  string   `json:duration`
	ID        primitive.ObjectID `bson:"_id"`
	Qpid      string             `json:"qpid"`
	Questions []string           `json:"questions"`
	UserPhone string             `json:"userphone"`
}
