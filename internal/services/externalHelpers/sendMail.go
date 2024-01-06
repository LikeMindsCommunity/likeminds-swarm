package externalHelpers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

func SendMail(userId string, mail_recipients []string, subject string, body string) {

	if len(mail_recipients) == 0 {
		return
	}

	headers := gin.H{
		"x-member-di": userId,
	}

	requestBody := gin.H{
		"subject":             subject,
		"mail_body":           body,
		"mail_recipient_list": mail_recipients,
	}

	respBytes, statusCode, err := GetRequestResponse(CaravanService, SendMailEndpoint, POSTRequestRawBody, headers, nil, requestBody)
	if err != nil {
		logging.Error(fmt.Sprintf("Error while sending mail to team: %s", err.Error()))
		return
	}

	if statusCode != 200 {
		logging.Error(fmt.Sprintf("Error while sending mail | statusCode: %d %s", statusCode, string(respBytes)))
	} else {
		logging.Info(fmt.Sprintf("Mail sent by user: %s to mails: %v with subject: %s", userId, mail_recipients, subject))
	}
}
