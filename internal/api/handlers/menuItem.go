package handlers

import (
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
)

func getMenuItem(menuItemName string) requests.MenuResponse {
	var menuId int
	var menuTitle string

	switch menuItemName {
	case constants.DeletePostMenuItemName:
		menuId = constants.DeletePostMenuItemId
		menuTitle = constants.DeletePostMenuItemTitle

	case constants.PinPostMenuItemName:
		menuId = constants.PinPostMenuItemId
		menuTitle = constants.PinPostMenuItemTitle

	case constants.UnpinPostMenuItemName:
		menuId = constants.UnpinPostMenuItemId
		menuTitle = constants.UnpinPostMenuItemTitle

	case constants.ReportPostMenuItemName:
		menuId = constants.ReportPostMenuItemId
		menuTitle = constants.ReportPostMenuItemTitle

	case constants.DeleteCommentMenuItemName:
		menuId = constants.DeleteCommentMenuItemId
		menuTitle = constants.DeleteCommentMenuItemTitle

	case constants.ReportCommentMenuItemName:
		menuId = constants.ReportCommentMenuItemId
		menuTitle = constants.ReportCommentMenuItemTitle
	}

	return requests.MenuResponse{
		Id:    menuId,
		Title: menuTitle,
	}
}

// Exposed Method to get Post Menu for Owner who are CMs also
func GetIsOwnerIsCmPostMenuItems(is_pinned bool) []requests.MenuResponse {
	menuItems := []requests.MenuResponse{getMenuItem(constants.DeletePostMenuItemName)}

	if is_pinned {
		menuItems = append(menuItems, getMenuItem(constants.UnpinPostMenuItemName))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.PinPostMenuItemName))
	}
	return menuItems
}

// Exposed Method to get Post Menu for Owner who is not a CM
func GetIsOwnerNotIsCmPostMenuItems() []requests.MenuResponse {
	return []requests.MenuResponse{getMenuItem(constants.DeletePostMenuItemName)}
}

// Exposed Method to get Post Menu for CMs who are not owners
func GetNotIsOwnerIsCmPostMenuItems(is_pinned bool) []requests.MenuResponse {
	menuItems := []requests.MenuResponse{}

	if is_pinned {
		menuItems = append(menuItems, getMenuItem(constants.UnpinPostMenuItemName))
	} else {
		menuItems = append(menuItems, getMenuItem(constants.PinPostMenuItemName))
	}

	menuItems = append(menuItems, getMenuItem(constants.DeletePostMenuItemName))
	return menuItems
}

// Exposed Method to get Post Menu for members
func GetNotIsOwnerNotIsCmPostMenuItems() []requests.MenuResponse {
	return []requests.MenuResponse{getMenuItem(constants.ReportPostMenuItemName)}
}

// Exposed Method to get Comment Menu for owner
func GetIsOwnerCommentMenuItems() []requests.MenuResponse {
	return []requests.MenuResponse{getMenuItem(constants.DeleteCommentMenuItemName)}
}

// Exposed Method to get Comment Menu for CM who is not owner
func GetNotIsOwnerIsCmCommentMenuItems() []requests.MenuResponse {
	return []requests.MenuResponse{getMenuItem(constants.DeleteCommentMenuItemName)}
}

// Exposed Method to get Comment Menu for members
func GetNotIsOwnerNotIsCmCommentMenuItems() []requests.MenuResponse {
	return []requests.MenuResponse{getMenuItem(constants.ReportCommentMenuItemName)}
}
