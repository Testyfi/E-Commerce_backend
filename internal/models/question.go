package models

type Question struct {
	Question string   `bson:"question"`
	Images   []string `bson:"images"`
	Type     int      `bson:"type"`
	Options  []string `bson:"options"`
}
