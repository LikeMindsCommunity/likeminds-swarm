package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ActivityRepository | Interface for Activity Repository
type ActivityRepository interface {
	Create(like *entities.Activity) (interface{}, error)
	Find(filter map[string]interface{}, filterOptions *options.FindOptions) ([]entities.Activity, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

// ActivityHelper | Interface for Activity Helper
type ActivityHelper interface {
	CreateActivityHelper(communityID int, actionBy []string, actionOn string, entityType constants.EntityType, entityID primitive.ObjectID,
		entityOwnerID string, action constants.ActivityAction, cta string, isRead bool) (interface{}, error)
	FindActivityHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Activity, error)
	UpdateActivityByIDHelper(activityID primitive.ObjectID, update map[string]interface{}) error
	CountActivityHelper(filter map[string]interface{}) (int64, error)
}
