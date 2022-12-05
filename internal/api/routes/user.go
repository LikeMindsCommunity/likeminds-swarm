package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
)

func UserRouter(routerGroup *gin.RouterGroup, activityHelper interfaces.ActivityHelper) {
	activityHandlers := handlers.NewActivityHandlers(activityHelper)

	postGroup := routerGroup.Group("user")
	postGroup.POST("/:user_id/activity", activityHandlers.ExternalCreateActivity)
}
