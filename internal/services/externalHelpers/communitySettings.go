package externalHelpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/cache"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/utils"
	"github.com/gin-gonic/gin"
)

// CommunitySetting | schema for community settings
type CommunitySetting struct {
	SettingType     string `json:"setting_type"  binding:"required"`
	SettingTitle    string `json:"setting_title"  binding:"required"`
	SettingSubTitle string `json:"setting_sub_title"  binding:"required"`
	IsEnabled       bool   `json:"enabled"  binding:"required"`
}

type CommunitySettingsResponse struct {
	Success           bool               `json:"success"`
	CommunitySettings []CommunitySetting `json:"community_settings"`
}

func fetchCommunitySettingsFromCache(cacheHelper cache.Helper, communityId int) []CommunitySetting {

	communitySettingsCacheKey := fmt.Sprintf(cache.CommunitySettingsCacheKey, communityId)
	communitySettingsCacheValue := cacheHelper.Get(communitySettingsCacheKey)

	if communitySettingsCacheValue.Val() == "" || communitySettingsCacheValue.Val() == "null" {
		logging.Error(fmt.Sprintf("Community settings not found in cache for communityId: %d", communityId))
		return nil
	}

	var communitySettings []CommunitySetting
	err := json.Unmarshal([]byte(communitySettingsCacheValue.Val()), &communitySettings)
	if err != nil {
		return nil
	}

	return communitySettings
}

func saveCommunitySettingsInCache(cacheHelper cache.Helper, communityId int, communitySettings []CommunitySetting) error {

	communitySettingsCacheKey := fmt.Sprintf(cache.CommunitySettingsCacheKey, communityId)
	parsedCommunitySettings, err := json.Marshal(communitySettings)
	if err != nil {
		return err
	}

	communitySettingsSet := cacheHelper.Set(communitySettingsCacheKey, parsedCommunitySettings, cache.CommunitySettingsCacheTTLInHours*time.Hour)
	if communitySettingsSet.Err() != nil {
		logging.Error(fmt.Sprintf("Error while Saving community settings in cache for communityId: %d, err: %v", communityId, err))
		return err
	}

	logging.Info(fmt.Sprintf("Saved community settings in cache for communityId: %d", communityId))
	return nil
}

// fetch community setting for application internal use
func getCommunitySettingsFromAPI(userId string, communityId int) ([]CommunitySetting, error) {

	headers := gin.H{
		"Content-Type": "application/json",
		"x-member-id":  userId,
	}

	// Params to be sent in the api/community/configurations
	params := map[string]string{
		ParamCommunityId: fmt.Sprintf("%d", communityId),
	}

	// Send request to internal service
	respBytes, statusCode, err := GetRequestResponse(CaravanService, FetchCommunitySettingsEndpoint, GETRequest, headers, params, nil)
	if err != nil || statusCode != http.StatusOK {
		return nil, err
	}

	// Parse response
	var csr CommunitySettingsResponse
	if err = json.Unmarshal(respBytes, &csr); err != nil {
		return nil, err
	}

	if len(csr.CommunitySettings) == 0 {
		return nil, errors.New("Community settings not found.")
	}

	return csr.CommunitySettings, nil
}

// method to fetch CommunitySettings
func fetchCommunitySettings(cacheHelper cache.Helper, userId string, communityId int) ([]CommunitySetting, error) {

	// fetch communitySettings from cache
	communitySettings := fetchCommunitySettingsFromCache(cacheHelper, communityId)
	if communitySettings == nil {

		// fetch from api if not found in cache
		communitySettings, err := getCommunitySettingsFromAPI(userId, communityId)
		if err != nil {
			return nil, err
		}

		// save in cache
		utils.SafeGo(func() { saveCommunitySettingsInCache(cacheHelper, communityId, communitySettings) })
	}

	return communitySettings, nil
}

// check is setting type is enabled in the community
func checkCommunitySettingEnabled(communitySettings []CommunitySetting, settingType string) bool {
	for _, setting := range communitySettings {
		if setting.SettingType == settingType {
			return setting.IsEnabled
		}
	}
	return false
}

// Exposed method to check if user connection setting is enabled for a community
func IsUserConnectionSettingEnabled(cacheHelper cache.Helper, userId string, communityId int) bool {

	communitySettings, err := fetchCommunitySettings(cacheHelper, userId, communityId)
	if err != nil {
		logging.Error(fmt.Sprintf("Error while fetching community settings, err: %v", err))
		return false
	}

	return checkCommunitySettingEnabled(communitySettings, UserConnectionSettingType)
}

// Expose method to get community ids list for which particluar setting is enabled
func GetCommunityIdsForCommunitySettingsEnabled(cacheHelper cache.Helper, communitySettingType string) []string {
	communityIdsList := []string{}

	communitySettingsAllKeys := cacheHelper.GetKeysFromPattern(cache.AllCommunitySettingsCacheKey)
	communitySettingKeys := communitySettingsAllKeys.Val()

	if communitySettingKeys == nil {
		logging.Error(fmt.Sprintf("Community settings not found in cache for all the communities"))
		return communityIdsList
	}

	re := regexp.MustCompile(CommunityIdFromCommunitySettingsRegex)

	for _, communitySettingKey := range communitySettingKeys {
		communityIdsMatchList := re.FindStringSubmatch(communitySettingKey)

		if len(communityIdsMatchList) == 2 {
			communityIdString := communityIdsMatchList[1]
			communityId, err := strconv.Atoi(communityIdString)
			if err == nil {
				communitySettings := fetchCommunitySettingsFromCache(cacheHelper, communityId)
				if checkCommunitySettingEnabled(communitySettings, communitySettingType) {
					communityIdsList = append(communityIdsList, communityIdString)
				}
			}
		}
	}

	return communityIdsList
}
