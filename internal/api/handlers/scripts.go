package handlers

import (
	"fmt"
	"time"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
)

func (handlers *FeedHandlers) IndexAllPostData() error {
	fmt.Println("starting IndexAllPostData script")
	startTime := time.Now()

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
		err = handlers.esHelper.IndexDocument(ParsePostIndexData(&postData), postData.ID.Hex(), constants.PostIndexName)
		if err != nil {
			log.Error(err.Error())
		}
	}
	fmt.Printf("script: IndexAllPostData, executed in %d\n", time.Since(startTime))

	return nil
}

func (handlers *FeedHandlers) IndexAllTopicData() error {
	fmt.Println("starting IndexAllTopicData script")
	startTime := time.Now()

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
		// index topic data in elastic search
		err = handlers.esHelper.IndexDocument(ParseTopicIndexData(handlers.postHelper, &topicData, true), topicData.ID.Hex(), constants.TopicIndexName)
		if err != nil {
			log.Error(err.Error())
		}
	}
	fmt.Printf("script: IndexAllTopicData, executed in %d\n", time.Since(startTime))

	return nil
}

func (handlers *FeedHandlers) IndexAllWidgetData() error {
	fmt.Println("starting IndexAllWidgetData script")
	startTime := time.Now()

	// delete widget index in elastic search
	err := handlers.esHelper.DeleteIndex(constants.WidgetIndexName)
	if err != nil {
		return err
	}

	// create widget index in elastic search
	err = handlers.esHelper.CreateIndex(constants.WidgetIndexName)
	if err != nil {
		return err
	}

	// fetch widget using helper method
	widgetResults, err := handlers.widgetHelper.FindWidgetHelper(gin.H{}, gin.H{})
	if err != nil {
		return err
	}

	for _, widgetData := range widgetResults {
		// insert widget data in elastic search
		err = handlers.esHelper.IndexDocument(ParseWidgetIndexData(&widgetData), widgetData.ID.Hex(), constants.WidgetIndexName)
		if err != nil {
			log.Error(err.Error())
		}
	}
	fmt.Printf("script: IndexAllWidgetData, executed in %d\n", time.Since(startTime))

	return nil
}

// function to insert community_id to all comments of a post
func (handlers *FeedHandlers) InsertCommunityIDToAllComments() error {
	fmt.Println("starting InsertCommunityIDToAllComments script")
	startTime := time.Now()

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
	fmt.Printf("script: InsertCommunityIDToAllComments, executed in %d\n", time.Since(startTime))

	return nil
}

// function to insert Post topics of all posts
func (handlers *FeedHandlers) BackfillPostTopicsInDB() error {
	fmt.Println("Starting BackfillPostTopicsInDB script!")
	startTime := time.Now()

	// fetch all comments
	postsResults, err := handlers.postHelper.FindPostHelper(gin.H{}, gin.H{})
	if err != nil {
		return err
	}

	for _, post := range postsResults {
		// Create new data in post topics
		if err := createOrUpdatePostTopics(handlers, post.ID.Hex(), false); err != nil {
			log.Error(fmt.Sprintf("Error while creating post topics data for post: %s, %s", post.ID.Hex(), err.Error()))
			continue
		}
	}
	fmt.Printf("script: BackfillPostTopicsInDB, executed in %d\n", time.Since(startTime))

	return nil
}

// function to insert default values for fields in topics collection
func (handlers *FeedHandlers) BackfillDefaultValuesForTopics() error {

	fmt.Println("Starting InsertDefaultValuesForTopics script!")
	startTime := time.Now()

	// fields to be updated
	fields := map[string]interface{}{
		"priority":          0,
		"is_searchable":     true,
		"parent_id":         primitive.NilObjectID,
		"parent_name":       "",
		"all_parent_ids":    nil,
		"level":             0,
		"widget_id":         primitive.NilObjectID,
		"total_child_count": 0,
	}

	// update fields with default values in topics collection if not exists
	for key, defaultValue := range fields {

		filter := bson.M{key: bson.M{"$exists": false}}
		update := bson.M{"$set": bson.M{key: defaultValue}}

		err := handlers.topicHelper.UpdateManyTopicsHelper(filter, update, false)
		if err != nil {
			log.Error(err)
		}
	}

	fmt.Printf("script: InsertDefaultValuesForTopics, executed in %d\n", time.Since(startTime))
	return nil
}
