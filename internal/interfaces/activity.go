package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityRepository interface {
	Create(like *entities.Activity) (interface{}, error)
	Find(filter map[string]interface{}) ([]entities.Activity, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
}

type ActivityHelper interface {
	CreateActivityHelper(action_by string, action_on []string, community_id int, entity_type string, entity_id primitive.ObjectID,
		action string, cta_data map[string]interface{}) (interface{}, error)
	FindActivityHelper(filter map[string]interface{}) ([]entities.Activity, error)
	UpdateActivityByIdHelper(activity_id primitive.ObjectID, update map[string]interface{}) error
}
