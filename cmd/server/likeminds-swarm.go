package main

import (
	"fmt"
	"os"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v7"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/middlewares"
	"github.com/nateshr/likeminds-swarm/internal/repositories"
	"github.com/nateshr/likeminds-swarm/internal/scripts"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/services/worker/distributor"
	"github.com/nateshr/likeminds-swarm/internal/services/worker/processor"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/api/routes"
	"github.com/nateshr/likeminds-swarm/internal/services/database"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/web"
)

const (
	AppVersion     string = "1.12.0" // Application Version
	GinPortAddress string = ":8080"  // Gin Port Address
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

// Internal Method to inject dependencies and get feed handlers
func injectDependenciesAndGetHandler(dbClient *mongo.Database, redisClient *redis.Client, esClient *elasticsearch.Client) *handlers.FeedHandlers {

	// Dependency injection of repositories
	postRepository := repositories.NewPostRepository(dbClient)
	pendingPostRepository := repositories.NewPendingPostRepository(dbClient)
	likeRepository := repositories.NewLikeRepository(dbClient)
	commentRepository := repositories.NewCommentRepository(dbClient)
	saveRepository := repositories.NewSaveRepository(dbClient)
	activityRepository := repositories.NewActivityRepository(dbClient)
	topicRepository := repositories.NewTopicRepository(dbClient)
	widgetRepository := repositories.NewWidgetRepository(dbClient)
	pollVotesRepository := repositories.NewPollVotesRepository(dbClient)
	connectionFeedRepository := repositories.NewConnectionFeedRepository(dbClient)

	// Dependency injection of Cache & ES
	cacheHelper := cache.NewCacheHelper(redisClient)
	esHelper := searchElastic.NewESHelper(esClient)

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
	feedTaskDistributor := distributor.NewTaskDistributor()

	// return feed handlers
	return handlers.NewFeedHandlers(likeHelper, commentHelper, postHelper, pendingPostHelper, saveHelper, activityHelper,
		topicHelper, widgetHepler, pollVotesHelper, connectionFeedHelper, esHelper, cacheHelper, feedTaskDistributor)
}

// Main Method
func main() {

	// Initiate clients
	redisClient := cache.InitRedis()
	dbClient := database.InitiateDB()
	esClient := searchElastic.InitiateES()

	// Inject dependencies and get feed handlers
	feedHandlers := injectDependenciesAndGetHandler(dbClient, redisClient, esClient)

	switch {
	// By default, run gin server
	case len(os.Args) == 1:

		// Initiate Gin Router
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
		logging.Fatal(router.Run(GinPortAddress))

	// runworker | Run background worker to process tasks
	case os.Args[1] == "runworker":

		feedTaskProcessor := processor.NewTaskProcessor(feedHandlers)

		// Run Background worker to process tasks
		logging.Fatal(feedTaskProcessor.Run())

	// runscript | Run script to perform some action
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
