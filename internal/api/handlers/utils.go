package handlers

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	OrderTypeAscending  int = 1
	OrderTypeDescending int = -1
	OrderTypeDefault    int = 0
)

// Internal Method to get pagination params in an API
func fetchPaginationParams(c *gin.Context) (int, int, error) {
	page, err := utils.ParseIntFromQueryParam(c.DefaultQuery("page", "0"), 0)
	if err != nil {
		return 0, 0, err
	}

	page_size, err := utils.ParseIntFromQueryParam(c.DefaultQuery("page_size", "0"), 0)
	if err != nil {
		return 0, 0, err
	}

	if page <= 0 {
		page = 1
	}

	if page_size <= 0 {
		page_size = 10
	}

	if page_size > 100 {
		return page, 0, fmt.Errorf("max page_size limit exceeded")
	}

	return page, page_size, nil
}

// Internal method to add sorting options to a map
func addSortingOptions(options map[string]interface{}, orderBy string, order int) map[string]interface{} {
	if order >= OrderTypeDefault {
		order = OrderTypeAscending
	} else {
		order = OrderTypeDescending
	}

	sortOpts, ok := options["$sort"]
	if ok {
		bsonSortOpts, ok := sortOpts.(bson.D)
		if ok {
			bsonSortOpts = append(bsonSortOpts, bson.E{Key: orderBy, Value: order})
			options["$sort"] = bsonSortOpts
		} else {
			options["$sort"] = bson.D{
				{Key: orderBy, Value: order},
			}
		}
	} else {
		options["$sort"] = bson.D{
			{Key: orderBy, Value: order},
		}
	}

	return options
}

// Internal Method to generate filter from page params from an API
func generatePageFilterOptions(c *gin.Context, sortKeyParam string, sortKeyOrderParam int) (map[string]interface{}, error) {
	// fetch pagination query params
	page, page_size, err := fetchPaginationParams(c)
	if err != nil {
		return nil, err
	}

	// page filter options
	filter_options := gin.H{
		"$skip":  page_size * (page - 1),
		"$limit": page_size,
	}

	sortKey := "created_at"

	if sortKeyParam != "" {
		sortKey = sortKeyParam
	}

	sortKeyOrder := OrderTypeDescending

	if sortKeyOrderParam != OrderTypeDefault {
		sortKeyOrder = sortKeyOrderParam
	}

	filter_options = addSortingOptions(filter_options, sortKey, sortKeyOrder)

	return filter_options, nil
}

// Internal Method to get tagged user Ids from a Text
func getTaggedUsers(text string) ([]string, error) {
	tagged_members := []string{}
	pattern, err := regexp.Compile("route://[member member_profile]+/([a-f0-9]{8}-?[a-f0-9]{4}-?4[a-f0-9]{3}-?[89ab][a-f0-9]{3}-?[a-f0-9]{12})")
	if err != nil {
		return nil, err
	}

	allSubstringMatches := pattern.FindAllStringSubmatch(text, -1)

	for _, occurance := range allSubstringMatches {
		tagged_members = append(tagged_members, occurance[1])
	}

	return tagged_members, nil
}

// Internal Method to fetch menu items for a user on an Entity
func getEntityMenuItems(entityType string, isCm bool, isOwner bool, isPinned bool,
	versionCode string, platformCode string, userId string, communityId int, cacheHelper cache.Helper,
	entityCreatorId string) []responses.MenuResponse {

	var output_menu_items []responses.MenuResponse
	var externalEntities externalHelpers.ExternalEntities

	isEditEnabled := utils.CheckVersion(utils.EditFeedEntityVersions, versionCode, platformCode)

	switch entityType {
	case constants.PostEntityType:
		// Get community configurations
		communityConfigurationResponse, _ := externalHelpers.GetCommunityConfigurations(cacheHelper, userId, communityId)
		isPostApprovalSettingEnabled := externalHelpers.IsPostApprovalNeeded(cacheHelper, userId, communityId)

		if communityConfigurationResponse != nil {
			externalEntities = externalHelpers.ExternalEntities{
				CommunityConfigurations:      communityConfigurationResponse.CommunityConfigurations,
				IsPostApprovalSettingEnabled: isPostApprovalSettingEnabled,
			}
		}

		// Get users list who are blocked by userId or blocked the userId
		blockUserValuesList, _ := externalHelpers.GetUserBlockList(cacheHelper, userId, communityId)

		isEntityOwnerBlocked := false
		if len(utils.GetSimilarBetweenArray([]string{entityCreatorId}, blockUserValuesList.BlockedUsers)) > 0 {
			isEntityOwnerBlocked = true
		}

		if isOwner && isCm {
			output_menu_items = GetIsOwnerIsCmPostMenuItems(isPinned, isEditEnabled, externalEntities)
		}

		if isOwner && !isCm {
			output_menu_items = GetIsOwnerNotIsCmPostMenuItems(isEditEnabled, externalEntities)
		}

		if !isOwner && isCm {
			output_menu_items = GetNotIsOwnerIsCmPostMenuItems(isPinned, isEditEnabled, externalEntities, isEntityOwnerBlocked)
		}

		if !isOwner && !isCm {
			output_menu_items = GetNotIsOwnerNotIsCmPostMenuItems(isEditEnabled, externalEntities, isEntityOwnerBlocked)
		}

	case constants.CommentEntityType:
		if isOwner {
			output_menu_items = GetIsOwnerCommentMenuItems(isEditEnabled, externalEntities)
		}

		if !isOwner && isCm {
			output_menu_items = GetNotIsOwnerIsCmCommentMenuItems(isEditEnabled, externalEntities)
		}

		if !isOwner && !isCm {
			output_menu_items = GetNotIsOwnerNotIsCmCommentMenuItems(isEditEnabled, externalEntities)
		}

	case constants.PendingPostEntityType:
		output_menu_items = GetPendingPostMenuItems(isEditEnabled, externalEntities)
	}

	return output_menu_items
}

// Internal Method to check if a number lies in Fibonacci Series
func checkIfFibonacciNumber(num int) bool {
	var n3, n1, n2 int = 0, 0, 1

	if num == n1 || num == n2 {
		return true
	}

	n3 = n1 + n2

	for {
		if n3 > num {
			break
		}

		if n3 == num {
			return true
		}
		n1 = n2
		n2 = n3
		n3 = n1 + n2
	}

	return false
}
