package externalHelpers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// Structure for Connection Object
type Connection struct {
	User1UUID string `json:"user1_uuid"`
	User2UUID string `json:"user2_uuid"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
}

// Structure for Connection Response
type ConnectionResponse struct {
	Success     bool                  `json:"success"`
	Connections []Connection          `json:"connections"`
	Users       map[string]MemberMeta `json:"users"`
}

// FetchUserConnections | fetch user connections for a give page and page size
func FetchUserConnectionsByPage(userId string, communityId int, page int, pageSize int) (bool, *ConnectionResponse) {
	headers := gin.H{
		"Content-Type": "application/json",
		"x-member-id":  userId,
	}

	// Params to be sent in the api/community_member/user_uuid/connection GET request
	params := map[string]string{
		ParamPage:        fmt.Sprintf("%d", page),
		ParamPageSize:    fmt.Sprintf("%d", pageSize),
		ParamCommunityId: fmt.Sprintf("%d", communityId),
	}

	// Endpoint generation
	endPoint := fmt.Sprintf(FetchUserConnectionsEndPoint, userId)

	//Send Request
	respBytes, _, err := GetRequestResponse(CaravanService, endPoint, GETRequest, headers, params, nil)
	if respBytes == nil {
		log.Error(fmt.Sprintf("FetchUserConnectionsByPage() - Error while fetching data from Caravan Service, %s", err.Error()))
		return false, nil
	}

	var connectionResponse ConnectionResponse
	if err := json.Unmarshal(respBytes, &connectionResponse); err != nil {
		//Internal unmarshal error
		return false, nil
	}

	return true, &connectionResponse
}
