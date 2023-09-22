package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ActivityRepository | Interface for Activity Repository
type ActivityRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOptions *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	UpdateAll(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

// ActivityHelper | Interface for Activity Helper
type ActivityHelper interface {
	CreateActivityHelper(communityID int, actionBy []string, actionOn string, entityType constants.EntityType, entityID primitive.ObjectID,
		entityOwnerID string, action constants.ActivityAction, cta string, isRead bool, isDeleted bool) (interface{}, error)
	FindActivityHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Activity, error)
	UpdateActivityByIDHelper(activityID primitive.ObjectID, update map[string]interface{}, shouldNotUpdateTimestamp bool, shouldPushActivityToCache bool) error
	CountActivityHelper(filter map[string]interface{}) (int64, error)
	DeleteActivityHelper(filter map[string]interface{}) error
	WarmupUserActivityFeedCache(communityID int, userID string) []entities.Activity
	PushActivitytoCache(activityID interface{})
	UpdateActivityInCache(activityID string)
}
