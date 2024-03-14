package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Widget Repository
type WidgetRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Delete(filter map[string]interface{}) error
	DeleteMany(filter map[string]interface{}) (int64, error)
	Count(filter map[string]interface{}) (int64, error)
}

// Interface for Widget Helper
type WidgetHelper interface {
	CreateWidgetHelper(createdByLM bool, parentEntityID string, parentEntityType string, metaData map[string]interface{},
		lmMeta map[string]interface{}, community_id int) (interface{}, error)
	FindWidgetHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Widget, error)
	UpdateWidgetByIdHelper(widgetId primitive.ObjectID, update map[string]interface{}) error
	DeleteWidgetByIdHelper(widgetId primitive.ObjectID) error
	DeleteWidgetsHelper(filter map[string]interface{}) (int64, error)
	CountWidgetHelper(filter map[string]interface{}) (int64, error)
}
