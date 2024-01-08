package handlers

import (
	"fmt"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
)

func (handlers *FeedHandlers) IndexAllPostData() error {
	// delete post index in elastic search
	err := handlers.esHelper.DeleteIndex(constants.PostIndexName)
	if err != nil {
		return err
	}

	// create post index in elastic search
	err = handlers.esHelper.CreateIndex(constants.PostIndexName)
	if err != nil {
		return err
	}

	// post filter data
	postFilterData := gin.H{
		"is_deleted": false,
	}

	// fetch post using helper method
	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, gin.H{})
	if err != nil {
		return err
	}

	for _, postData := range postResults {
		// insert post data in elastic search
		err = handlers.esHelper.InsertDocument(ParsePostIndexData(&postData), postData.ID.Hex(), constants.PostIndexName)
		if err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

func (handlers *FeedHandlers) IndexAllTopicData() error {
	// delete topic index in elastic search
	err := handlers.esHelper.DeleteIndex(constants.TopicIndexName)
	if err != nil {
		return err
	}

	// create topic index in elastic search
	err = handlers.esHelper.CreateIndex(constants.TopicIndexName)
	if err != nil {
		return err
	}

	// fetch topic using helper method
	topicResults, err := handlers.topicHelper.FindTopicHelper(gin.H{}, gin.H{})
	if err != nil {
		return err
	}

	for _, topicData := range topicResults {
		// insert topic data in elastic search
		err = handlers.esHelper.InsertDocument(ParseTopicIndexData(&topicData), topicData.ID.Hex(), constants.TopicIndexName)
		if err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

// function to insert community_id to all comments of a post
func (handlers *FeedHandlers) InsertCommunityIDToAllComments() error {

	// fetch all comments
	commentResults, err := handlers.commentHelper.FindCommentHelper(gin.H{}, gin.H{})
	if err != nil {
		return err
	}

	for _, comment := range commentResults {
		postId := comment.PostId

		// fetch post using helper method
		postData, err := handlers.postHelper.FindPostHelper(gin.H{"_id": postId}, gin.H{})
		if err != nil || len(postData) == 0 {
			log.Error(fmt.Sprintf("Post not found for comment id: %s", comment.ID.Hex()))
			continue
		}

		// comment update data
		commentUpdateData := gin.H{
			"$set": gin.H{
				"community_id": postData[0].CommunityId,
			},
		}

		// update comment data
		err = handlers.commentHelper.UpdateCommentByIdHelper(comment.ID, commentUpdateData)
		if err != nil {
			log.Error(fmt.Sprintf("Error while updating comment with id: %s", comment.ID.Hex()))
			continue
		}
	}

	return nil
}
