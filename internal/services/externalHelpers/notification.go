package externalHelpers

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Exposed Method to send Notification using Caravan Service API
func SendNotification(member_ids []string, title string, sub_title string, route string, community_id int, category string, subCategory string,
	platform_code string, version_code string) {

	titleNotificationVersion := utils.CheckVersionInverted(utils.NotificationVersions, version_code, platform_code)

	if !titleNotificationVersion && title == "" && sub_title != "" {
		title = sub_title
		sub_title = ""
	}

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
		"Content-Type":    ContentTypeHeader,
		"x-member-id":     SwarmServiceMemberIdHeader,
		"x-platform-code": platform_code,
		"x-version-code":  version_code,
	}

	//Send Request
	respBytes, _, err := GetRequestResponse(CaravanService, SendNotificationEndPoint, POSTRequestRawBody, headers, nil, postBody)
	if respBytes == nil {
		log.Fatalf("An Error Occured %v", err)
	}

	// Printing output
	sb := string(respBytes)
	fmt.Println(sb)
}
