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
