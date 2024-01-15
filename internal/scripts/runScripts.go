package scripts

import (
	"fmt"
	"time"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

func RunScripts(handlers *handlers.FeedHandlers) {
	// indexPostData(handlers)
	// indexTopicData(handlers)
	// indexWidgetData(handlers)
	// addCommunityIdToComments(handlers)
}

func indexPostData(handlers *handlers.FeedHandlers) {
	err := handlers.IndexAllPostData()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running indexPostData: %s", err.Error()))
		return
	}
}

func indexTopicData(handlers *handlers.FeedHandlers) {
	fmt.Println("starting function indexTopicData")
	startTime := time.Now()
	err := handlers.IndexAllTopicData()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running indexTopicData: %s", err.Error()))
		return
	}
	fmt.Println("function indexTopicData completed in ", time.Since(startTime))
}

func indexWidgetData(handlers *handlers.FeedHandlers) {
	fmt.Println("starting function indexWidgetData")
	startTime := time.Now()

	err := handlers.IndexAllWidgetData()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running indexWidgetData: %s", err.Error()))
		return
	}
	fmt.Println("function indexWidgetData completed in ", time.Since(startTime))
}

func addCommunityIdToComments(handlers *handlers.FeedHandlers) {
	err := handlers.InsertCommunityIDToAllComments()
	if err != nil {
		log.Error(fmt.Sprintf("Scripts: Error running addCommunityIdToComments: %s", err.Error()))
		return
	}
	log.Info("Scripts: addCommunityIdToAllComments completed successfully")
}
