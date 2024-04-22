package externalHelpers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/utils"
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
	State           int           `json:"state"`
	SDKClientInfo   SDKClientInfo `json:"sdk_client_info"`
	IsDeleted       bool          `json:"is_deleted"`
	QuestionAnswers []interface{} `json:"question_answers"`
}

// Structure for Member Meta Response
type MemberMetaResponse struct {
	Success bool         `json:"success"`
	Members []MemberMeta `json:"members"`
}

// FetchMemberMeta | fetch member meta for sent id
func FetchMemberMeta(memberIds []string, userId string, communityId int) (bool, *MemberMetaResponse) {

	// Call API with Api Version
	headers := gin.H{
		utils.HeadersMemberId:   userId,
		utils.HeadersSdkSource:  utils.SdkSourceFeed,
		utils.HeadersApiVersion: utils.FetchMembersMetaApiVersion,
	}

	paramMemberIds, _ := json.Marshal(memberIds)

	//Params to be sent in the api/community_member/fetch_access request
	params := map[string]string{
		ParamMemberIds:   fmt.Sprintf("%v", string(paramMemberIds)),
		ParamCommunityId: fmt.Sprintf("%d", communityId),
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

// FetchMemberMetaMap | fetch member meta for sent ids and return map
func FetchMemberMetaMap(member_ids []string, userId string, communityId int) (bool, map[string]MemberMeta) {

	memberMetaMap := map[string]MemberMeta{}

	//Fetch Member Meta
	success, memberMetaResponse := FetchMemberMeta(member_ids, userId, communityId)
	if !success {
		return false, memberMetaMap
	}

	//Create Member Meta Map
	for _, memberMeta := range memberMetaResponse.Members {
		memberMetaMap[memberMeta.UserUniqueId] = memberMeta
	}

	return true, memberMetaMap
}
