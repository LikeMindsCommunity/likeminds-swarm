package helpers

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (helper *likeHelper) CreateLikeHelper(entity_type string, entity_id primitive.ObjectID, liked_by string) (interface{}, error) {
	like := entities.NewLike(entity_type, entity_id, liked_by)
	like_id, err := helper.likeRepository.Create(&like)

	return like_id, err
}

func (helper *likeHelper) FindLikeHelper(filter map[string]interface{}) ([]entities.Like, error) {
	results, err := helper.likeRepository.Find(filter)

	return results, err
}

func (helper *likeHelper) UpdateLikeHelper(filter map[string]interface{}, update map[string]interface{}) error {
	err := helper.likeRepository.Update(filter, update)

	return err
}

func (helper *likeHelper) CountLikeHelper(filter map[string]interface{}) (int64, error) {
	count, err := helper.likeRepository.Count(filter)

	return count, err
}

type likeHelper struct {
	likeRepository interfaces.LikeRepository
}

func NewLikeHelper(likeRepository interfaces.LikeRepository) interfaces.LikeHelper {
	return &likeHelper{
		likeRepository: likeRepository,
	}
}
