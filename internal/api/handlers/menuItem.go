package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func getMenuItem(menuItemName string, externalEntities externalHelpers.ExternalEntities) responses.MenuResponse {
	var menuId int
	var menuTitle string
	var postFeedMetadataValues string

	communityConfiguration, _ := externalHelpers.GetCommunityConfigurationAgainstType(externalEntities.CommunityConfigurations,
		externalHelpers.FeedMetadataCommunityConfigurationType)

	feedMetadataPostVariableValue, isFetched := communityConfiguration.Value[externalHelpers.PostCommunityConfigurationKey]

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
func GetIsOwnerIsCmPostMenuItems(isPinned bool, isHidden bool, isEditCheck bool, externalEntities externalHelpers.ExternalEntities,
) []responses.MenuResponse {

	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditPostMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeletePostMenuItemName, externalEntities))

	if isPinned {
		menuItems = append(menuItems, getMenuItem(constants.UnpinPostMenuItemName, externalEntities))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.PinPostMenuItemName, externalEntities))
	}

	if isHidden {
		menuItems = append(menuItems, getMenuItem(constants.UnHidePostMenuItemName, externalEntities))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.HidePostMenuItemName, externalEntities))
	}

	return menuItems
}

// Exposed Method to get Post Menu for Owner who is not a CM
func GetIsOwnerNotIsCmPostMenuItems(isEditCheck bool, externalEntities externalHelpers.ExternalEntities) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditPostMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeletePostMenuItemName, externalEntities))

	return menuItems
}

// Exposed Method to get Post Menu for CMs who are not owners
func GetNotIsOwnerIsCmPostMenuItems(isPinned bool, isHidden bool, isEditCheck bool, externalEntities externalHelpers.ExternalEntities, isEntityOwnerBlocked bool) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditPostMenuItemName, externalEntities))
	}

	if isPinned {
		menuItems = append(menuItems, getMenuItem(constants.UnpinPostMenuItemName, externalEntities))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.PinPostMenuItemName, externalEntities))
	}

	if isHidden {
		menuItems = append(menuItems, getMenuItem(constants.UnHidePostMenuItemName, externalEntities))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.HidePostMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeletePostMenuItemName, externalEntities))

	// if !isEntityOwnerBlocked {
	// 	// Block user menu item
	// 	menuItems = append(menuItems, getMenuItem(constants.BlockUserMenuItemName, externalEntities))
	// }

	return menuItems
}

// Exposed Method to get Post Menu for members
func GetNotIsOwnerNotIsCmPostMenuItems(isEditCheck bool, externalEntities externalHelpers.ExternalEntities, isEntityOwnerBlocked bool) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{getMenuItem(constants.ReportPostMenuItemName, externalEntities)}

	// if !isEntityOwnerBlocked {
	// 	// Block user menu item
	// 	menuItems = append(menuItems, getMenuItem(constants.BlockUserMenuItemName, externalEntities))
	// }

	return menuItems

}

// Exposed Method to get Comment Menu for owner
func GetIsOwnerCommentMenuItems(isEditCheck bool, externalEntities externalHelpers.ExternalEntities) []responses.MenuResponse {

	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditCommentMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeleteCommentMenuItemName, externalEntities))

	return menuItems
}

// Exposed Method to get Comment Menu for CM who is not owner
func GetNotIsOwnerIsCmCommentMenuItems(isEditCheck bool, externalEntities externalHelpers.ExternalEntities) []responses.MenuResponse {

	menuItems := []responses.MenuResponse{}

	if isEditCheck {
		menuItems = append(menuItems, getMenuItem(constants.EditCommentMenuItemName, externalEntities))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeleteCommentMenuItemName, externalEntities))

	return menuItems
}

// Exposed Method to get Comment Menu for members
func GetNotIsOwnerNotIsCmCommentMenuItems(isEditCheck bool, externalEntities externalHelpers.ExternalEntities) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{getMenuItem(constants.ReportCommentMenuItemName, externalEntities)}

	return menuItems
}

// Exposed Method to get Pending Post Menu for creators
func GetPendingPostMenuItems(isEditCheck bool, externalEntities externalHelpers.ExternalEntities) []responses.MenuResponse {
	menuItems := []responses.MenuResponse{
		getMenuItem(constants.EditPendingPostMenuItemName, externalEntities),
		getMenuItem(constants.DeletePendingPostMenuItemName, externalEntities),
	}

	return menuItems
}
