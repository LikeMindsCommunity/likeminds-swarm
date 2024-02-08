package main

import (
	"fmt"
	"os"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/middlewares"
	"github.com/nateshr/likeminds-swarm/internal/scripts"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/api/routes"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/repositories"
	"github.com/nateshr/likeminds-swarm/internal/services/database"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/web"
)

// Internal Method to enable Cors in the server
func enableCors() cors.Config {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders(
		"x-member-id",
		"x-platform-code",
		"x-version-code",
		"x-sdk-source",
		"x-device-id",
		"x-api-key",
	)
	return config
}

// Internal Method to initiate Gin module in the server
func initGin() *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Enable CORS
	router.Use(cors.New(enableCors()))

	// Enable Logging Middleware
	router.Use(middlewares.LoggingMiddleware())

	return router
}

// Internal Method to initiate the server
func main() {
	var AppVersion string = "1.12.0"

	redisClient := cache.InitRedis()
	db := database.InitiateDB()
	es := searchElastic.InitiateES()

	// Dependency injection of repositories
	postRepository := repositories.NewPostRepository(db)
	pendingPostRepository := repositories.NewPendingPostRepository(db)
	likeRepository := repositories.NewLikeRepository(db)
	commentRepository := repositories.NewCommentRepository(db)
	saveRepository := repositories.NewSaveRepository(db)
	activityRepository := repositories.NewActivityRepository(db)
	topicRepository := repositories.NewTopicRepository(db)
	widgetRepository := repositories.NewWidgetRepository(db)
	pollVotesRepository := repositories.NewPollVotesRepository(db)
	connectionFeedRepository := repositories.NewConnectionFeedRepository(db)

	// Dependency injection of Cache & ES
	cacheHelper := cache.NewCacheHelper(redisClient)
	esHelper := searchElastic.NewESHelper(es)

	// Dependency injection of helpers
	postHelper := helpers.NewPostHelper(postRepository)
	pendingPostHelper := helpers.NewPendingPostHelper(pendingPostRepository)
	likeHelper := helpers.NewLikeHelper(likeRepository)
	commentHelper := helpers.NewCommentHelper(commentRepository)
	saveHelper := helpers.NewSaveHelper(saveRepository)
	activityHelper := helpers.NewActivityHelper(activityRepository, cacheHelper)
	topicHelper := helpers.NewTopicHelper(topicRepository)
	widgetHepler := helpers.NewWidgetHelper(widgetRepository)
	pollVotesHelper := helpers.NewPollVotesHelper(pollVotesRepository)
	connectionFeedHelper := helpers.NewConnectionFeedHelper(connectionFeedRepository)

	// initiate task distributor for background tasks
	feedTaskDistributor := handlers.NewTaskDistributor()

	// New feed Handler
	feedHandlers := handlers.NewFeedHandlers(likeHelper, commentHelper, postHelper, pendingPostHelper, saveHelper,
		activityHelper, topicHelper, widgetHepler, pollVotesHelper, connectionFeedHelper, esHelper, cacheHelper, feedTaskDistributor)

	switch {

	// By default, run the server
	case len(os.Args) == 1:

		router := initGin()

		router.GET("", web.Home)
		routerGroup := router.Group("/")

		// Routes
		routes.BaseRouter(routerGroup, feedHandlers)
		routes.PostRouter(routerGroup, feedHandlers)
		routes.UserRouter(routerGroup, feedHandlers)
		routes.FeedRouter(routerGroup, feedHandlers)
		routes.TopicRouter(routerGroup, feedHandlers)
		routes.WidgetRouter(routerGroup, feedHandlers)
		routes.PollRouter(routerGroup, feedHandlers)

		logging.Info(fmt.Sprintf("Main: application version: %s", AppVersion))

		// Run gin server
		logging.Fatal(router.Run(":8080"))

	// Run background worker to process tasks
	case os.Args[1] == "runworker":

		// Run Background Worker to process tasks
		logging.Fatal(runBackgroundWorker(feedHandlers))

	// Run Scripts
	case os.Args[1] == "runscript":
		if len(os.Args) == 3 {
			scripts.RunScripts(feedHandlers, os.Args[2])
		} else {
			logging.Fatal("please provide a valid script name to run")
		}
		os.Exit(0) // TODO: Check if needed else remove

	default:
		logging.Fatal("Invalid args")
	}
}

// Internal Method to run background worker
func runBackgroundWorker(handler *handlers.FeedHandlers) error {

	// initiate redis client
	redisOpt := asynq.RedisClientOpt{
		Addr: environment.GoDotEnvVariable("ASYNQ_BROKER_ADDRESS"),
	}

	// initiate task processor
	taskProcessor := handlers.NewTaskProcessor(redisOpt, handler)

	// run the task processor
	return taskProcessor.Run()
}
