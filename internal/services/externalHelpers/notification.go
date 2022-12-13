package externalHelpers

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func SendNotification(member_ids []string, title string, sub_title string, route string, community_id int) {
	postBody := gin.H{
		"community_id": community_id,
		"member_ids":   member_ids,
		"message_payload": gin.H{
			"title":     title,
			"sub_title": sub_title,
			"route":     route,
		},
	}

	headers := gin.H{
		"Content-Type": "application/json",
		"x-member-id":  "swarm-service",
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
