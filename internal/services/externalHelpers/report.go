package externalHelpers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

func pushReport(headers gin.H, postBody gin.H) error {

	if headers == nil || postBody == nil {
		return fmt.Errorf("headers or postBody is nil")
	}

	// Send request
	respBytes, statusCode, err := GetRequestResponse(CaravanService, PushReportEndpoint, POSTRequestRawBody, headers,
		nil, postBody)
	if err != nil {
		logging.Error(fmt.Sprintf("error while pushing report: %s", err.Error()))
		return err
	}

	if statusCode != http.StatusOK {
		logging.Error(fmt.Sprintf("error while pushing report | statusCode: %d , Response:  %s", statusCode, string(respBytes)))
		return fmt.Errorf("error while pushing report:  %s", string(respBytes))
	}

	return nil
}

// SendPendingPostForReview | Exposed method for pushing pending post for review
func SendPendingPostForReview(userId string, communityId int, pendingPostId string) error {

	headers := gin.H{
		"x-member-id": userId,
	}

	postBody := gin.H{
		"entity_id":    pendingPostId,
		"entity_type":  enums.EntityTypePendingPost,
		"accused_uuid": userId,
		"community_id": fmt.Sprint(communityId),
	}

	return pushReport(headers, postBody)
}
