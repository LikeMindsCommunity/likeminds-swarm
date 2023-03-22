package scripts

import (
	"log"

	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

func RunScripts(handlers *handlers.FeedHandlers) {
	// indexPostData(handlers)
	// addCommunityIdToComments(handlers)
}

func indexPostData(handlers *handlers.FeedHandlers) {
	err := handlers.IndexAllPostData()
	if err != nil {
		log.Fatalf("Scripts: Error running indexPostData: %s", err.Error())
		return
	}
}

func addCommunityIdToComments(handlers *handlers.FeedHandlers) {
	err := handlers.InsertCommunityIDToAllComments()
	if err != nil {
		log.Println("Scripts: Error running addCommunityIdToComments: ", err.Error())
		return
	}
	log.Println("Scripts: addCommunityIdToAllComments completed successfully")
}
