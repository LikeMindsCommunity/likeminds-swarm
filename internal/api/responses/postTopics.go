package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type PostIdsBasedonTopics struct {
	PostIDs []primitive.ObjectID `json:"post_ids" bson:"post_ids"`
}
