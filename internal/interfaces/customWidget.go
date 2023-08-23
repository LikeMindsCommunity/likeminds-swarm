package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for CustomWidget Repository
type CustomWidgetRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

// Interface for CustomWidget Helper
type CustomWidgetHelper interface {
	CreateCustomWidgetHelper(createdByLM bool, parentEntityID string, parentEntityType string, metaData map[string]interface{},
		community_id int) (interface{}, error)
	FindCustomWidgetHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.CustomWidget,
		error)
	UpdateCustomWidgetByIdHelper(customWidgetId primitive.ObjectID, update map[string]interface{}) error
	CountCustomWidgetHelper(filter map[string]interface{}) (int64, error)
}
