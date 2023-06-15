package scripts

import (
	"fmt"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

func RunScripts(handlers *handlers.FeedHandlers) {
	// indexPostData(handlers)
	// addCommunityIdToComments(handlers)
}

func indexPostData(handlers *handlers.FeedHandlers) {
	err := handlers.IndexAllPostData()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running indexPostData: %s", err.Error()))
		return
	}
}

func addCommunityIdToComments(handlers *handlers.FeedHandlers) {
	err := handlers.InsertCommunityIDToAllComments()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running addCommunityIdToComments: %s", err.Error()))
		return
	}
	log.Info("Scripts: addCommunityIdToAllComments completed successfully")
}
