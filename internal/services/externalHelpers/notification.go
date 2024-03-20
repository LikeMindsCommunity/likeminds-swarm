package externalHelpers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"

	"github.com/gin-gonic/gin"
)

// Exposed Method to send Notification using Caravan Service API
func SendNotification(member_ids []string, title string, sub_title string, route string, community_id int, category string, subCategory string,
	platform_code string, version_code string) {

	postBody := gin.H{
		"community_id": community_id,
		"member_ids":   member_ids,
		"message_payload": gin.H{
			"title":     title,
			"sub_title": sub_title,
			"route":     route,
		},
		"category": gin.H{
			"category":    category,
			"subcategory": subCategory,
		},
	}

	headers := gin.H{
		"Content-Type":            ContentTypeHeader,
		utils.HeadersMemberId:     SwarmServiceHeader,
		utils.HeadersPlatformCode: platform_code,
		utils.HeadersVersionCode:  version_code,
		utils.HeadersSdkSource:    utils.SdkSourceFeed,
	}

	//Send Request
	respBytes, _, err := GetRequestResponse(CaravanService, SendNotificationEndPoint, POSTRequestRawBody, headers, nil, postBody)
	if respBytes == nil {
		logging.Error(fmt.Sprintf("An Error Occured %v", err))
	}

	// Printing output
	sb := string(respBytes)
	fmt.Println(sb)
}
