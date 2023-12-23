package handlers

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func TotalMarks(s string)int  {
	if(s=="Jee Mains"){return 300}
    total:=0
	filter := bson.D{{"subject_tags", s}}
	//, {"type", "Single Correct"}
	// Count the number of documents that match the filter
	count, err :=questionCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	total=int(count)*4
	filter = bson.D{{"subject_tags", s}, {"type", "Single Correct"}}
	count, err =questionCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	return total-int(count)
}
