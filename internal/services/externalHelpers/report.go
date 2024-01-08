package externalHelpers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

func pushReport(userId string, communityId int, postBody gin.H) error {

	headers := gin.H{
		"x-member-id":    userId,
		"x-community-id": fmt.Sprint(communityId),
	}

	// Send request
	respBytes, statusCode, err := GetRequestResponse(CaravanService, PushReportEndpoint, POSTRequestRawBody, headers,
		nil, postBody)
	if err != nil {
		logging.Error(fmt.Errorf("error while pushing report: %s", err.Error()))
		return err
	}

	if statusCode != http.StatusOK {
		logging.Error(fmt.Errorf("error while pushing report | statusCode: %d , Response:  %s", statusCode, string(respBytes)))
		return fmt.Errorf("error while pushing report:  %s", string(respBytes))
	}

	return nil
}

// SendPendingPostForReview | Exposed method for pushing pending post for review
func SendPendingPostForReview(userId string, communityId int, pendingPostId string) error {

	postBody := gin.H{
		"entity_id":       pendingPostId,
		"entity_type":     enums.EntityTypePendingPost,
		"entity_owner_id": userId,
	}

	return pushReport(userId, communityId, postBody)
}
