package externalHelpers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

// SDKClientInfo | defines sdk client info object schema
type SDKClientInfo struct {
	User         int    `json:"user"`
	UserUniqueID string `json:"user_unique_id"`
	UUID         string `json:"uuid"`
	Community    int    `json:"community"`
	WidgetId     string `json:"widget_id"`
}

// Structure for Member Meta Object
type MemberMeta struct {
	Id              int           `json:"id"`
	Name            string        `json:"name"`
	ImageUrl        string        `json:"image_url"`
	UserUniqueId    string        `json:"user_unique_id"`
	UUID            string        `json:"uuid"`
	IsGuest         bool          `json:"is_guest"`
	CustomTitle     string        `json:"custom_title"`
	SDKClientInfo   SDKClientInfo `json:"sdk_client_info"`
	QuestionAnswers []interface{} `json:"question_answers"`
}

// Structure for Member Meta Response
type MemberMetaResponse struct {
	Success bool         `json:"success"`
	Members []MemberMeta `json:"members"`
}

// FetchMemberMeta | fetch member meta for sent ids
func FetchMemberMeta(member_ids []string, user_id string, community_id int) (bool, *MemberMetaResponse) {
	headers := gin.H{
		"Content-Type": "application/json",
		"x-member-id":  user_id,
	}

	paramMemberIds, _ := json.Marshal(member_ids)

	//Params to be sent in the api/community_member/fetch_access request
	params := map[string]string{
		ParamMemberIds:   fmt.Sprintf("%v", string(paramMemberIds)),
		ParamCommunityId: fmt.Sprintf("%d", community_id),
	}

	//Send Request
	respBytes, _, _ := GetRequestResponse(CaravanService, FetchMembersMetaEndPoint, GETRequest, headers, params, nil)
	if respBytes == nil {
		return false, nil
	}

	var membersMetaResponse MemberMetaResponse
	if err := json.Unmarshal(respBytes, &membersMetaResponse); err != nil {
		//Internal unmarshal error
		return false, nil
	}

	return true, &membersMetaResponse
}
