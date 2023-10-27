package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/nateshr/likeminds-swarm/internal/entities"
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
		err = handlers.esHelper.InsertDocument(context.Background(), ParsePostIndexData(&postData), postData.ID.Hex(), constants.PostIndexName)
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
		err = handlers.esHelper.InsertDocument(context.Background(), ParseTopicIndexData(&topicData), topicData.ID.Hex(), constants.TopicIndexName)
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

type UserData struct {
	Cuuid           string `json:"client_uuid"`
	CommunityId     int    `json:"community_id"`
	LMuuid          string `json:"lm_uuid"`
	UserId          int    `json:"user_id"`
	ChatInteraction bool   `json:"chat_interaction"`
	FeedInteraction bool   `json:"feed_interaction"`
}

type UserInteractionMap map[string][]UserData

func (handlers *FeedHandlers) ParseJsonAndGetUserInteraction() error {

	// load json file
	jsonFile, err := os.Open("internal/scripts/prod_users_data.json")
	if err != nil {
		return err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read json file
	byteValue, _ := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	// unmarshal json file
	var usersData map[string]interface{}
	err = json.Unmarshal(byteValue, &usersData)
	if err != nil {

		return err
	}

	finalUserResponse := UserInteractionMap{}

	// iterate over users data
	for uuid, users := range usersData {

		// fmt.Println("For UUID ->", uuid)

		usersData := []UserData{}

		users, err := json.Marshal(users)
		if err != nil {
			return err
		}

		err = json.Unmarshal(users, &usersData)
		if err != nil {
			return err
		}

		usersInteractionArray := []UserData{}

		FeedInteraction, feedUserIndex := false, -1
		ChatInteraction := false

		// iterate over all users of the same UUID
		for i, userData := range usersData {

			// fetch user interaction
			_, feed_interaction, err := fetchUserInteraction(handlers, userData.LMuuid, userData.CommunityId)
			if err != nil {
				return err
			}

			if feed_interaction {
				userData.FeedInteraction = true
				// fmt.Println("User Interacted -> ", userInteraction)
				FeedInteraction = true
				feedUserIndex = i
			}

			if userData.ChatInteraction {
				ChatInteraction = true
			}

			usersInteractionArray = append(usersInteractionArray, userData)
		}

		// to find uuids with 2 different interaction account
		if FeedInteraction && ChatInteraction {

			for i, user := range usersInteractionArray {
				if user.ChatInteraction && (i != feedUserIndex) {
					fmt.Println(" UUID found with 2 different interaction account -> ", uuid)
				}
			}
		}

		finalUserResponse[uuid] = usersInteractionArray
	}

	// Dump UserInteractionMap to json file
	userInteractionJson, err := json.Marshal(finalUserResponse)
	if err != nil {
		return err
	}

	err = os.WriteFile("internal/scripts/prod_user_interaction.json", userInteractionJson, 0644)
	if err != nil {
		return err
	}

	return nil

}

func fetchUserInteraction(handlers *FeedHandlers, userId string, communityId int) (map[string]interface{}, bool, error) {

	if_interacted := false

	// fetch user posts
	userPosts, _, err := fetchUserPosts(handlers, userId, communityId)
	if err != nil {
		return nil, if_interacted, err
	}

	// fetch user comments
	userComments, _, err := fetchUserComments(handlers, userId, communityId)
	if err != nil {
		return nil, if_interacted, err
	}

	// fetch user likes
	userLikes, _, err := fetchUserLikes(handlers, userId, communityId)
	if err != nil {
		return nil, if_interacted, err
	}

	userInteraction := map[string]interface{}{
		"posts":    userPosts,
		"comments": userComments,
		"likes":    userLikes,
	}

	if len(userPosts) > 0 || len(userComments) > 0 || len(userLikes) > 0 {
		if_interacted = true
	}

	return userInteraction, if_interacted, nil
}

func fetchUserPosts(handlers *FeedHandlers, userId string, communityId int) ([]entities.Post, int, error) {

	// post filter data
	postFilterData := gin.H{
		"user_id":      userId,
		"community_id": communityId,
		"is_deleted":   false,
	}

	// fetch post using helper method
	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, gin.H{})
	if err != nil {
		return nil, 0, err
	}

	return postResults, len(postResults), nil
}

func fetchUserComments(handlers *FeedHandlers, userId string, communityId int) ([]entities.Comment, int, error) {

	// comment filter data
	commentFilterData := gin.H{
		"user_id":      userId,
		"community_id": communityId,
		"is_deleted":   false,
	}

	// fetch comment using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, gin.H{})
	if err != nil {
		return nil, 0, err
	}

	return commentResults, len(commentResults), nil
}

func fetchUserLikes(handlers *FeedHandlers, userId string, communityId int) ([]entities.Like, int, error) {

	// like filter data
	likeFilterData := gin.H{
		"user_id":      userId,
		"community_id": communityId,
	}

	// fetch like using helper method
	likeResults, err := handlers.likeHelper.FindLikeHelper(likeFilterData, gin.H{})
	if err != nil {
		return nil, 0, err
	}

	return likeResults, len(likeResults), nil
}
