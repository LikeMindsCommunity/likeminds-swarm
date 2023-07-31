package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create a Like
func (helper *likeHelper) CreateLikeHelper(entity_type string, entity_id primitive.ObjectID, liked_by string) (interface{}, error) {
	like := entities.NewLike(entity_type, entity_id, liked_by)
	like_id, err := helper.likeRepository.Create(&like)

	return like_id, err
}

// Exposed Helper Method to Find Likes
func (helper *likeHelper) FindLikeHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Like, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.likeRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.Like
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Likes
func (helper *likeHelper) UpdateLikeByIdHelper(like_id primitive.ObjectID, update map[string]interface{}) error {
	set_data := gin.H{}

	if _, ok := update["$set"]; ok {
		set_data = update["$set"].(gin.H)
	}
	set_data["updated_at"] = time.Now()
	update["$set"] = set_data

	err := helper.likeRepository.Update(gin.H{"_id": like_id}, update)

	return err
}

// Exposed Helper Method to Fetch Likes Count
func (helper *likeHelper) CountLikeHelper(filter map[string]interface{}) (int64, error) {
	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.likeRepository.Count(filter)

	return count, err
}

// Exposed Helper Method to perform Aggregation on Likes
func (helper *likeHelper) AggregateLikeHelper(query []map[string]interface{}) (interface{}, error) {
	for _, value := range query {
		if matchGroup, ok := value["$match"]; ok {
			err := convertHexIdsToObjectIds(matchGroup.(gin.H), []string{"_id", "entity_id"})
			if err != nil {
				return nil, err
			}
		}
	}

	results, err := helper.likeRepository.Aggregate(query)

	return results, err
}

// Structure for Like Helper
type likeHelper struct {
	likeRepository interfaces.LikeRepository
}

// Exposed Method to create New Like Helper
func NewLikeHelper(likeRepository interfaces.LikeRepository) interfaces.LikeHelper {
	return &likeHelper{
		likeRepository: likeRepository,
	}
}
