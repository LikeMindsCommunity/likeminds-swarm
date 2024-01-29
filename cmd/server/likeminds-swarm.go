package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/scripts"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

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

var (
	redisClient *redis.Client
	router      *gin.Engine
)

// Internal Method to initiate the server
func main() {
	var AppVersion string = "1.12.0"

	redisClient = cache.InitRedis()
	db := database.InitiateDB()
	es := searchElastic.InitiateES()

	cacheHelper := cache.NewCacheHelper(redisClient)

	redisOpt := asynq.RedisClientOpt{
		Addr: "0.0.0.0:6381",
	}

	taskDistributor := handlers.NewRedisTaskDistributor(redisOpt)

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

	// Dependency injection of elasticSearch Helper
	esHelper := searchElastic.NewESHelper(es)

	// New feed Handler
	feedHandlers := handlers.NewFeedHandlers(likeHelper, commentHelper, postHelper, pendingPostHelper, saveHelper,
		activityHelper, topicHelper, widgetHepler, pollVotesHelper, connectionFeedHelper, esHelper, cacheHelper, taskDistributor)

	switch {
	case len(os.Args) == 1:
		//run server here
		fmt.Println("----------run server here")
		initGin()
		router.Use(cors.New(enableCors()))
		router.Use(LoggingMiddleware())

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

		log.Info(fmt.Sprintf("Main: application version: %s", AppVersion))
		log.Fatal(router.Run(":8090"))
		break
	case os.Args[1] == "runworker":
		// run worker here
		fmt.Println("----------run worker here")
		runTaskProcessor(redisOpt, feedHandlers)
		// select {}
		break
	case os.Args[1] == "runscript":
		fmt.Println("----------run script here")
		fmt.Println(os.Args[2])
		// Run Scripts
		scripts.RunScripts(feedHandlers, os.Args[2])
		os.Exit(0)
		break
	default:
		//TODO:
		// throw invalid args error
	}

	//TODO:
	//cases:
	// 1-> routes
	// 2-> script?
	// 3-> worker

	// config -> Builder pattern
	// opts := Config.Builder().option1("value").build() -> 5 retries
	// specific task -> updatedOpts := opts.toBuilder().option2("value2").build()

	// default options
	// override -> overriden otherwise default values
}

// Internal Method to initiate Gin module in the server
func initGin() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.Default()
}

// Method to process API request to log
func processRequest(c *gin.Context) interface{} {
	requestBodyData := gin.H{}

	// Reading request body
	requestBody, err := io.ReadAll(c.Request.Body)

	// Updating request body after read
	c.Request.Body = io.NopCloser(bytes.NewReader(requestBody))

	// Unmarshalling request body
	if err == nil {
		_ = json.Unmarshal(requestBody, &requestBodyData)
	}

	return gin.H{
		"host":         c.Request.Host,
		"absolute_uri": c.Request.RequestURI,
		"method":       c.Request.Method,
		"headers":      c.Request.Header,
		"body":         requestBodyData,
	}
}

// responseBodyWriter | Custom Response Writer
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write | Custom Write Method for responseBodyWriter
func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// LoggingMiddleware will log the request and response of API
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/" {

			c.Next()

		} else {

			data := gin.H{}

			// Starting time
			startTime := time.Now()

			// Implementing custom response body writer in the context
			w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
			c.Writer = w

			// Updating Request Data
			data["request"] = processRequest(c)

			// Processing request
			c.Next()

			// End Time
			endTime := time.Now()

			response := gin.H{}
			statusCode := c.Writer.Status()

			// Unmarshalling Request Response
			_ = json.Unmarshal(w.body.Bytes(), &response)

			// Updating Request Response
			data["response"] = gin.H{
				"http_response_code": statusCode,
				"content":            response,
			}

			if statusCode < http.StatusBadRequest {
				data["response"].(gin.H)["content"] = gin.H{}
			}

			// Updating Request Meta Data
			data["meta"] = gin.H{
				"latency":   endTime.Sub(startTime),
				"client_ip": c.ClientIP(),
			}

			// Marshalling the final Data
			marshelledData, _ := json.Marshal(data)

			if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
				// Logging the generated request data as Info
				log.Info(string(marshelledData))
			} else {
				// Logging the generated request data as Error
				log.Error(string(marshelledData))
			}

			c.Next()
		}
	}
}

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

func runTaskProcessor(redisOpt asynq.RedisClientOpt, handler *handlers.FeedHandlers) {
	taskProcessor := handlers.NewRedisTaskProcessor(redisOpt, handler)
	fmt.Println("Starting task processor")
	err := taskProcessor.Run()
	if err != nil {
		fmt.Println("Failed to start task processor")
	}
}
