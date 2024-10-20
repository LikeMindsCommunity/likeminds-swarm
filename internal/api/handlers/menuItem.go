package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func getMenuItem(menuItemName string, externalEntities *externalHelpers.ExternalEntities) responses.MenuResponse {

	var menuId int
	var menuTitle string
	var postFeedMetadataValues string

	feedMetadatConfig, _ := externalHelpers.GetCommunityConfigurationAgainstType(externalEntities.CommunityConfigurations,
		externalHelpers.FeedMetadataCommunityConfigurationType)

	feedMetadataPostVariableValue, isFetched := feedMetadatConfig.Value[externalHelpers.PostCommunityConfigurationKey]

	if !isFetched {
		feedMetadataPostVariableValue = externalHelpers.DefaultPostVariableValue
	}

	postFeedMetadataValues = utils.CapitalizeFirstLetter(feedMetadataPostVariableValue.(string))

	switch menuItemName {
	case constants.DeletePostMenuItemName:
		menuId = constants.DeletePostMenuItemId
		menuTitle = fmt.Sprintf(constants.DeletePostMenuItemTitle, postFeedMetadataValues)

	case constants.PinPostMenuItemName:
		menuId = constants.PinPostMenuItemId
		menuTitle = fmt.Sprintf(constants.PinPostMenuItemTitle, postFeedMetadataValues)

	case constants.UnpinPostMenuItemName:
		menuId = constants.UnpinPostMenuItemId
		menuTitle = fmt.Sprintf(constants.UnpinPostMenuItemTitle, postFeedMetadataValues)

	case constants.ReportPostMenuItemName:
		menuId = constants.ReportPostMenuItemId
		menuTitle = constants.ReportPostMenuItemTitle

	case constants.DeleteCommentMenuItemName:
		menuId = constants.DeleteCommentMenuItemId
		menuTitle = constants.DeleteCommentMenuItemTitle

	case constants.ReportCommentMenuItemName:
		menuId = constants.ReportCommentMenuItemId
		menuTitle = constants.ReportCommentMenuItemTitle

	case constants.EditPostMenuItemName:
		menuId = constants.EditPostMenuItemId
		menuTitle = fmt.Sprintf(constants.EditPostMenuItemTitle, postFeedMetadataValues)

	case constants.EditCommentMenuItemName:
		menuId = constants.EditCommentMenuItemId
		menuTitle = constants.EditCommentMenuItemTitle

	case constants.EditPendingPostMenuItemName:
		menuId = constants.EditPendingPostMenuItemId
		menuTitle = constants.EditPendingPostMenuItemTitle

	case constants.DeletePendingPostMenuItemName:
		menuId = constants.DeletePendingPostMenuItemId
		menuTitle = constants.DeletePendingPostMenuItemTitle

	case constants.BlockUserMenuItemName:
		menuId = constants.BlockUserMenuItemId
		menuTitle = constants.BlockUserMenuItemTitle
	case constants.HidePostMenuItemName:
		menuId = constants.HidePostMenuItemId
		menuTitle = fmt.Sprintf(constants.HidePostMenuItemTitle, postFeedMetadataValues)
	case constants.UnHidePostMenuItemName:
		menuId = constants.UnHidePostMenuItemId
		menuTitle = fmt.Sprintf(constants.UnHidePostMenuItemTitle, postFeedMetadataValues)
	}

	return responses.MenuResponse{
		Id:    menuId,
		Title: menuTitle,
	}
}

// Exposed Method to get Post Menu for Owner who are CMs also
func GetIsOwnerIsCmPostMenuItems(isPinned bool, isHidden bool, isEditCheck bool, externalEntities *externalHelpers.ExternalEntities,
) []responses.MenuResponse {

	menuItems := []responses.MenuResponse{}

	menuItemsConfig := GetMenuItemsConfig(externalEntities.CommunityConfigurations)

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditPostMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeletePostMenuItemName, externalEntities))

	if isPinned {
		menuItems = append(menuItems, getMenuItem(constants.UnpinPostMenuItemName, externalEntities))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.PinPostMenuItemName, externalEntities))
	}

	// If hide post menu item is enabled, append appropriate menu item
	if menuItemsConfig[constants.HidePostMenuItemConfig] {
		if isHidden {
			menuItems = append(menuItems, getMenuItem(constants.UnHidePostMenuItemName, externalEntities))
		} else {
			menuItems = append(menuItems, getMenuItem(constants.HidePostMenuItemName, externalEntities))
		}
	}

	return menuItems
}

// Exposed Method to get Post Menu for Owner who is not a CM
func GetIsOwnerNotIsCmPostMenuItems(isEditCheck bool, externalEntities *externalHelpers.ExternalEntities) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditPostMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeletePostMenuItemName, externalEntities))

	return menuItems
}

// Exposed Method to get Post Menu for CMs who are not owners
func GetNotIsOwnerIsCmPostMenuItems(isPinned bool, isHidden bool, isEditCheck bool, externalEntities *externalHelpers.ExternalEntities, isEntityOwnerBlocked bool) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{}

	menuItemsConfig := GetMenuItemsConfig(externalEntities.CommunityConfigurations)

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditPostMenuItemName, externalEntities))
	}

	if isPinned {
		menuItems = append(menuItems, getMenuItem(constants.UnpinPostMenuItemName, externalEntities))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.PinPostMenuItemName, externalEntities))
	}

	if menuItemsConfig[constants.HidePostMenuItemConfig] {
		if isHidden {
			menuItems = append(menuItems, getMenuItem(constants.UnHidePostMenuItemName, externalEntities))
		} else {
			menuItems = append(menuItems, getMenuItem(constants.HidePostMenuItemName, externalEntities))
		}
	}

	menuItems = append(menuItems, getMenuItem(constants.DeletePostMenuItemName, externalEntities))

	// if !isEntityOwnerBlocked {
	// 	// Block user menu item
	// 	menuItems = append(menuItems, getMenuItem(constants.BlockUserMenuItemName, externalEntities))
	// }

	return menuItems
}

// Exposed Method to get Post Menu for members
func GetNotIsOwnerNotIsCmPostMenuItems(isEditCheck bool, externalEntities *externalHelpers.ExternalEntities, isEntityOwnerBlocked bool) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{getMenuItem(constants.ReportPostMenuItemName, externalEntities)}

	// if !isEntityOwnerBlocked {
	// 	// Block user menu item
	// 	menuItems = append(menuItems, getMenuItem(constants.BlockUserMenuItemName, externalEntities))
	// }

	return menuItems

}

// Exposed Method to get Comment Menu for owner
func GetIsOwnerCommentMenuItems(isEditCheck bool, externalEntities *externalHelpers.ExternalEntities) []responses.MenuResponse {

	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditCommentMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeleteCommentMenuItemName, externalEntities))

	return menuItems
}

// Exposed Method to get Comment Menu for CM who is not owner
func GetNotIsOwnerIsCmCommentMenuItems(isEditCheck bool, externalEntities *externalHelpers.ExternalEntities) []responses.MenuResponse {

	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditCommentMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeleteCommentMenuItemName, externalEntities))

	return menuItems
}

// Exposed Method to get Comment Menu for members
func GetNotIsOwnerNotIsCmCommentMenuItems(isEditCheck bool, externalEntities *externalHelpers.ExternalEntities) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{getMenuItem(constants.ReportCommentMenuItemName, externalEntities)}

	return menuItems
}

// Exposed Method to get Pending Post Menu for creators
func GetPendingPostMenuItems(isEditCheck bool, externalEntities *externalHelpers.ExternalEntities) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{
		getMenuItem(constants.EditPendingPostMenuItemName, externalEntities),
		getMenuItem(constants.DeletePendingPostMenuItemName, externalEntities),
	}

	return menuItems
}

// Internal Method to fetch menu items for a user on an Entity
func getEntityMenuItems(entityType string, isCm bool, isOwner bool, isPinned bool, isHidden bool,
	versionCode string, platformCode string, userId string, communityId int, cacheHelper cache.Helper, entityCreatorId string,
) []responses.MenuResponse {

	var output_menu_items []responses.MenuResponse
	var externalEntities externalHelpers.ExternalEntities

	isEditEnabled := utils.CheckVersion(utils.EditFeedEntityVersions, versionCode, platformCode)

	switch entityType {
	case constants.PostEntityType:
		// Get community configurations
		communityConfigurationResponse, _ := externalHelpers.GetCommunityConfigurations(cacheHelper, userId, communityId)

		if communityConfigurationResponse != nil {
			externalEntities = externalHelpers.ExternalEntities{
				CommunityConfigurations: communityConfigurationResponse.CommunityConfigurations,
			}
		}

		// Get users list who are blocked by userId or blocked the userId
		blockUserValuesList, _ := externalHelpers.GetUserBlockList(cacheHelper, userId, communityId)

		isEntityOwnerBlocked := false
		if len(utils.GetSimilarBetweenArray([]string{entityCreatorId}, blockUserValuesList.BlockedUsers)) > 0 {
			isEntityOwnerBlocked = true
		}

		if isOwner && isCm {
			output_menu_items = GetIsOwnerIsCmPostMenuItems(isPinned, isHidden, isEditEnabled, &externalEntities)
		}

		if isOwner && !isCm {
			output_menu_items = GetIsOwnerNotIsCmPostMenuItems(isEditEnabled, &externalEntities)
		}

		if !isOwner && isCm {
			output_menu_items = GetNotIsOwnerIsCmPostMenuItems(isPinned, isHidden, isEditEnabled, &externalEntities, isEntityOwnerBlocked)
		}

		if !isOwner && !isCm {
			output_menu_items = GetNotIsOwnerNotIsCmPostMenuItems(isEditEnabled, &externalEntities, isEntityOwnerBlocked)
		}

	case constants.CommentEntityType:
		if isOwner {
			output_menu_items = GetIsOwnerCommentMenuItems(isEditEnabled, &externalEntities)
		}

		if !isOwner && isCm {
			output_menu_items = GetNotIsOwnerIsCmCommentMenuItems(isEditEnabled, &externalEntities)
		}

		if !isOwner && !isCm {
			output_menu_items = GetNotIsOwnerNotIsCmCommentMenuItems(isEditEnabled, &externalEntities)
		}

	case constants.PendingPostEntityType:
		output_menu_items = GetPendingPostMenuItems(isEditEnabled, &externalEntities)
	}

	return output_menu_items
}

// Exposed method to fetch MenuItemsConfig from community configurations
func GetMenuItemsConfig(communityConfigurations []externalHelpers.CommunityConfiguration) map[string]bool {

	menuItemsConfig := map[string]bool{
		constants.HidePostMenuItemConfig: false,
	}

	if len(communityConfigurations) == 0 {
		logging.Error("No community configurations found when fetching menu Items config")
		return menuItemsConfig
	}

	feedSettings, err := externalHelpers.GetCommunityConfigurationAgainstType(communityConfigurations, externalHelpers.FeedSettingsCommunityConfigurations)
	if err != nil {
		logging.Error("Error when fetching feedSettings against configurations: ", err.Error())
		return menuItemsConfig
	}

	config, ok := feedSettings.Value[externalHelpers.FeedSettingsMenuItemConfigKey]
	if ok {
		hidePostValue, ok := config.(map[string]interface{})[constants.HidePostMenuItemConfig]
		if ok {
			menuItemsConfig[constants.HidePostMenuItemConfig] = hidePostValue.(bool)
		}
	}

	return menuItemsConfig
}
