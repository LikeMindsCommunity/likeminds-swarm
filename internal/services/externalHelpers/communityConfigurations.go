package externalHelpers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

type CommunityConfirgurationResponse struct {
	Success                 bool                     `json:"success"`
	CommunityConfigurations []CommunityConfiguration `json:"community_configurations"`
}

func GetCommunityConfigurationAgainstType(communityConfigurations []CommunityConfiguration, communityConfigurationType string) (CommunityConfiguration, error) {

	if len(communityConfigurations) > 0 {

		for _, communityConfiguration := range communityConfigurations {

			if communityConfiguration.Type == communityConfigurationType {
				return communityConfiguration, nil
			}
		}
	}
	return CommunityConfiguration{}, nil
}

// Exposed Method to get Community ID from API Key
func GetCommunityConfigurations(cacheHelper cache.Helper, userId string, communityId int) (*CommunityConfirgurationResponse, error) {
	var communityConfigurationResponse CommunityConfirgurationResponse

	// Get data from cache
	cacheCommunityConfigurationsKey := fmt.Sprintf(cache.CommunityConfigurationsKey, communityId)
	communityConfigurationsCacheValue := cacheHelper.Get(cacheCommunityConfigurationsKey)

	if communityConfigurationsCacheValue.Val() != "" && communityConfigurationsCacheValue.Val() != "null" {
		err := json.Unmarshal([]byte(communityConfigurationsCacheValue.Val()), &communityConfigurationResponse.CommunityConfigurations)

		if err != nil {
			return nil, err
		}

		return &communityConfigurationResponse, nil
	}

	headers := gin.H{
		"Content-Type": "application/json",
		"x-member-id":  userId,
	}

	// Params to be sent in the api/community/configurations
	params := map[string]string{
		ParamCommunityId: fmt.Sprintf("%d", communityId),
	}

	//Send Request
	respBytes, _, err := GetRequestResponse(CaravanService, FetchCommunityConfigurations, GETRequest, headers, params, nil)

	if respBytes == nil {
		//If API fails or any other error
		return nil, err
	}

	if err := json.Unmarshal(respBytes, &communityConfigurationResponse); err != nil {
		//Internal unmarshal error
		return nil, err
	}

	// Save data to cache
	logging.Info(fmt.Sprintf("Saving the community configurations in cache for %d", communityId))
	cacheCommunityConfigurationsBytesValue, _ := json.Marshal(communityConfigurationResponse.CommunityConfigurations)
	cacheHelper.Set(cacheCommunityConfigurationsKey, cacheCommunityConfigurationsBytesValue, 6*time.Hour)

	return &communityConfigurationResponse, nil
}

func GetDefaultOrDbCommunityConfiguration(cacheHelper cache.Helper, userId string, communityId int) string {
	var postFeedMetadataValues string = DefaultFeedMetadataPostVariableValue

	communityConfigurationResponse, _ := GetCommunityConfigurations(cacheHelper, userId, communityId)

	if communityConfigurationResponse != nil {
		externalEntities := ExternalEntities{
			communityConfigurationResponse.CommunityConfigurations,
		}

		communityConfiguration, _ := GetCommunityConfigurationAgainstType(externalEntities.CommunityConfigurations,
			FeedMetadataCommunityConfigurationType)

		feedMetadataPostVariableValue, isFetched := communityConfiguration.Value[PostCommunityConfigurationKey]

		if isFetched {
			postFeedMetadataValues = feedMetadataPostVariableValue.(string)
		}
	}

	return postFeedMetadataValues
}
