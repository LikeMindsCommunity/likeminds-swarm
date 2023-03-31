package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/api/routes"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/repositories"
	"github.com/nateshr/likeminds-swarm/internal/services/database"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/web"
)

var (
	router *gin.Engine
)

// Internal Method to initiate the server
func main() {
	var AppVersion string = "0.5.0"
	environment.LoadGoDotEnv()

	initGin()
	db := database.InitiateDB()
	es := searchElastic.InitiateES()

	router.Use(cors.New(enableCors()))

	router.GET("", web.Home)

	routerGroup := router.Group("/")

	// Dependency injection of repositories
	postRepository := repositories.NewPostRepository(db)
	likeRepository := repositories.NewLikeRepository(db)
	commentRepository := repositories.NewCommentRepository(db)
	saveRepository := repositories.NewSaveRepository(db)
	activityRepository := repositories.NewActivityRepository(db)

	// Dependency injection of helpers
	postHelper := helpers.NewPostHelper(postRepository)
	likeHelper := helpers.NewLikeHelper(likeRepository)
	commentHelper := helpers.NewCommentHelper(commentRepository)
	saveHelper := helpers.NewSaveHelper(saveRepository)
	activityHelper := helpers.NewActivityHelper(activityRepository)

	// Dependency injection of elasticSearch Helper
	esHelper := searchElastic.NewESHelper(es)

	// New feed Handler
	feedHandlers := handlers.NewFeedHandlers(likeHelper, commentHelper, postHelper, saveHelper, activityHelper, esHelper)

	// Routes
	routes.BaseRouter(routerGroup, feedHandlers)
	routes.PostRouter(routerGroup, feedHandlers)
	routes.UserRouter(routerGroup, feedHandlers)
	routes.FeedRouter(routerGroup, feedHandlers)

	// Run Scripts
	// scripts.RunScripts(feedHandlers)

	log.Printf("Main: application version: %s", AppVersion)
	log.Fatal(router.Run(":8080"))
}

// Internal Method to initiate Gin module in the server
func initGin() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.Default()
}

// Internal Method to enable Cors in the server
func enableCors() cors.Config {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders(
		"x-member-id",
		"x-platform-code",
		"x-version-code",
		"x-device-id",
		"x-api-key",
	)
	return config
}
