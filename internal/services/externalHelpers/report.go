package externalHelpers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/enums"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/gin-gonic/gin"
)

type PushReportResponse struct {
	Success  bool `json:"success"`
	ReportID int  `json:"report_id"`
}

func pushReport(headers gin.H, postBody gin.H) (int, error) {
	var reportId int

	if headers == nil || postBody == nil {
		return reportId, fmt.Errorf("headers or postBody is nil")
	}

	// Send request
	respBytes, statusCode, err := GetRequestResponse(CaravanService, PushReportEndpoint, POSTRequestRawBody, headers,
		nil, postBody)
	if err != nil {
		logging.Error(fmt.Sprintf("error while pushing report: %s", err.Error()))
		return reportId, err
	}

	if statusCode != http.StatusOK {
		logging.Error(fmt.Sprintf("error while pushing report | statusCode: %d , Response:  %s", statusCode, string(respBytes)))
		return reportId, fmt.Errorf("error while pushing report:  %s", string(respBytes))
	}

	var pushReportResponse PushReportResponse

	if err := json.Unmarshal(respBytes, &pushReportResponse); err != nil {
		//Internal unmarshal error
		return reportId, err
	}

	return pushReportResponse.ReportID, nil
}

// CreatePendingPostReport | Exposed method for pushing pending post in report
func CreatePendingPostReport(userId string, communityId int, pendingPostId string) (int, error) {

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
