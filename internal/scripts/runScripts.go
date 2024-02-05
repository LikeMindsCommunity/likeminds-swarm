package scripts

import (
	"fmt"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

func RunScripts(handlers *handlers.FeedHandlers, scriptName string) {
	switch scriptName {
	case "indexPostData":
		indexPostData(handlers)
		break
	case "indexTopicData":
		indexTopicData(handlers)
		break
	case "indexWidgetData":
		indexWidgetData(handlers)
		break
	case "addCommunityIdToComments":
		addCommunityIdToComments(handlers)
		break
	default:
		log.Fatal("Script not found")
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
