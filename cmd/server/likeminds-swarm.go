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
	"github.com/nateshr/likeminds-swarm/internal/web"
)

var (
	router *gin.Engine
)

func main() {
	var AppVersion string = "0.1.0"
	environment.LoadGoDotEnv()

	initGin()
	db := database.InitiateDB()
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

	// New feed Handler
	feedHandlers := handlers.NewFeedHandlers(likeHelper, commentHelper, postHelper, saveHelper, activityHelper)

	// Routes
	routes.BaseRouter(routerGroup, feedHandlers)
	routes.PostRouter(routerGroup, feedHandlers)
	routes.UserRouter(routerGroup, feedHandlers)
	routes.FeedRouter(routerGroup, feedHandlers)

	log.Printf("application version: %s", AppVersion)
	log.Fatal(router.Run(":8080"))
}

func initGin() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.Default()
}

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
