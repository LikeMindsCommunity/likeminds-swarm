package scripts

import (
	"log"

	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

func RunScripts(handlers *handlers.FeedHandlers) {
	indexPostData(handlers)
}

func indexPostData(handlers *handlers.FeedHandlers) {
	err := handlers.IndexAllPostData()
	if err != nil {
		log.Fatalf("Scripts: Error running indexPostData: %s", err.Error())
		return
	}
}
