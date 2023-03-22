package handlers

import (
	"context"
	"log"

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
	post_filter_data := gin.H{
		"is_deleted": false,
	}

	// fetch post using helper method
	post_results, err := handlers.postHelper.FindPostHelper(post_filter_data, gin.H{})
	if err != nil {
		return err
	}

	for _, post_data := range post_results {
		// insert post data in elastic search
		err = handlers.esHelper.InsertDocument(context.Background(), ParsePostIndexData(&post_data), post_data.ID.Hex(), constants.PostIndexName)
		if err != nil {
			log.Print(err.Error())
		}
	}

	return nil
}

// function to insert community_id to all comments of a post
func (handlers *FeedHandlers) InsertCommunityIDToAllComments() error {

	// fetch all comments
	comment_results, err := handlers.commentHelper.FindCommentHelper(gin.H{}, gin.H{})
	if err != nil {
		return err
	}

	for _, comment := range comment_results {
		post_id := comment.PostId

		// fetch post using helper method
		post_data, err := handlers.postHelper.FindPostHelper(gin.H{"_id": post_id}, gin.H{})
		if err != nil || len(post_data) == 0 {
			log.Println("Post not found for comment id: ", comment.ID.Hex())
			continue
		}

		// comment update data
		comment_update_data := gin.H{
			"$set": gin.H{
				"community_id": post_data[0].CommunityId,
			},
		}

		// update comment data
		err = handlers.commentHelper.UpdateCommentByIdHelper(comment.ID, comment_update_data)
		if err != nil {
			log.Println("Error while updating comment with id: ", comment.ID.Hex())
			continue
		}
	}

	return nil
}
