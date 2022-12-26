package helpers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (helper *likeHelper) CreateLikeHelper(entity_type string, entity_id primitive.ObjectID, liked_by string) (interface{}, error) {
	like := entities.NewLike(entity_type, entity_id, liked_by)
	like_id, err := helper.likeRepository.Create(&like)

	return like_id, err
}

func (helper *likeHelper) FindLikeHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Like, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	results, err := helper.likeRepository.Find(filter, &fOpts)

	return results, err
}

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

func (helper *likeHelper) CountLikeHelper(filter map[string]interface{}) (int64, error) {
	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.likeRepository.Count(filter)

	return count, err
}

func (helper *likeHelper) AggregateLikeHelper(query []interface{}) (interface{}, error) {
	for _, value := range query {
		if matchGroup, ok := value.(gin.H)["$match"]; ok {
			err := convertHexIdsToObjectIds(matchGroup.(gin.H), []string{"_id", "entity_id"})
			if err != nil {
				return nil, err
			}
		}
	}

	results, err := helper.likeRepository.Aggregate(query)

	return results, err
}

type likeHelper struct {
	likeRepository interfaces.LikeRepository
}

func NewLikeHelper(likeRepository interfaces.LikeRepository) interfaces.LikeHelper {
	return &likeHelper{
		likeRepository: likeRepository,
	}
}
