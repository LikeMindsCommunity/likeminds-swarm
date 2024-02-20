package externalHelpers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

type NSFWConfigurations struct {
	Enabled       bool    `json:"enabled"`
	InferdoApiKey string  `json:"inferdo_api_key"`
	CutoffScore   float64 `json:"cutoff_score"`
	ErrorStatus   string  `json:"error_status"`
}

type UniversalFeedConfigurations struct {
	CommentSortOn    string `json:"comment_sort_order_key"`
	CommentSortOrder string `json:"comment_sort_order"`
	CommentCount     int    `json:"comment_count"`
}

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

// Exposed Method to fetch community configurations from cache and if not found then from API
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
	cacheHelper.Set(cacheCommunityConfigurationsKey, cacheCommunityConfigurationsBytesValue, CommunityConfigurationsCacheTTLInHours*time.Hour)

	return &communityConfigurationResponse, nil
}

func GetFeedPostVariableOrDefault(cacheHelper cache.Helper, userId string, communityId int) string {
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

func GetUniversalFeedConfigurationsData(cacheHelper cache.Helper, userId string, communityId int) *UniversalFeedConfigurations {
	// Default Universal Feed Configurations
	universalFeedConfigurations := UniversalFeedConfigurations{}

	// Get community configurations
	communityConfigurationResponse, _ := GetCommunityConfigurations(cacheHelper, userId, communityId)

	if communityConfigurationResponse != nil {
		externalEntities := ExternalEntities{
			communityConfigurationResponse.CommunityConfigurations,
		}

		feedMetatdataCommunityConfig, _ := GetCommunityConfigurationAgainstType(externalEntities.CommunityConfigurations,
			FeedMetadataCommunityConfigurationType)

		// Universal feed metadata configurations
		universalFeedMetadata, isFetched := feedMetatdataCommunityConfig.Value[UniversalFeedCommunityConfigurationKey]

		if isFetched {
			feedMetatdataCommunityConfigBytes, _ := json.Marshal(universalFeedMetadata)
			json.Unmarshal(feedMetatdataCommunityConfigBytes, &universalFeedConfigurations)
		}
	}

	return &universalFeedConfigurations
}

// Exposed Helper method to fetch NSFW Filtering Configurations for a community
func GetNSFWConfigurationsOrDefault(cacheHelper cache.Helper, userId string, communityId int) (bool, *NSFWConfigurations) {

	// Default NSFW Configurations
	nsfwConfigurations := NSFWConfigurations{
		Enabled:       false,
		InferdoApiKey: "",
		CutoffScore:   0.8,
		ErrorStatus:   "",
	}

	communityConfigurationResponse, _ := GetCommunityConfigurations(cacheHelper, userId, communityId)

	if communityConfigurationResponse != nil {
		externalEntities := ExternalEntities{
			communityConfigurationResponse.CommunityConfigurations,
		}

		communityConfiguration, err := GetCommunityConfigurationAgainstType(externalEntities.CommunityConfigurations,
			NSFWFilteringCommunityConfigurationType)
		if err != nil {
			return nsfwConfigurations.Enabled, &nsfwConfigurations
		}

		if communityConfiguration.Type == NSFWFilteringCommunityConfigurationType {

			bytes, _ := json.Marshal(communityConfiguration.Value)
			json.Unmarshal(bytes, &nsfwConfigurations)

		}

	}

	return nsfwConfigurations.Enabled, &nsfwConfigurations

}
