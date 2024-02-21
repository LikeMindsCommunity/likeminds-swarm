package scripts

import (
	"fmt"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

func RunScripts(handlers *handlers.FeedHandlers, scriptName string) {
	switch scriptName {

	// Run the script to index all post data
	case "indexPostData":
		indexPostData(handlers)

	// Run the script to index all topic data
	case "indexTopicData":
		indexTopicData(handlers)

	// Run the script to index all widget data
	case "indexWidgetData":
		indexWidgetData(handlers)

	// Run the script to add community id to all comments
	case "addCommunityIdToComments":
		addCommunityIdToComments(handlers)

	// If the script is not found
	default:
		log.Fatal(fmt.Sprintf(`Scripts: Script '%s' not found`, scriptName))
	}
}

func indexPostData(handlers *handlers.FeedHandlers) {
	err := handlers.IndexAllPostData()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running indexPostData: %s", err.Error()))
		return
	}
	log.Info("Scripts: indexPostData completed successfully")
}

func indexTopicData(handlers *handlers.FeedHandlers) {
	err := handlers.IndexAllTopicData()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running indexTopicData: %s", err.Error()))
		return
	}
	log.Info("Scripts: indexTopicData completed successfully")
}

func indexWidgetData(handlers *handlers.FeedHandlers) {
	err := handlers.IndexAllWidgetData()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running indexWidgetData: %s", err.Error()))
		return
	}
	log.Info("Scripts: indexWidgetData completed successfully")
}

func addCommunityIdToComments(handlers *handlers.FeedHandlers) {
	err := handlers.InsertCommunityIDToAllComments()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running addCommunityIdToComments: %s", err.Error()))
		return
	}
	log.Info("Scripts: addCommunityIdToAllComments completed successfully")
}
