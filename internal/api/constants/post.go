package constants

const PostEntityType string = "post"
const (
	ImageWidget int = iota + 1
	VideoWidget
	DocumentWidget
	LinkWidget
)

const (
	DeletePostMenuItem = "Delete Post"
	PinPostMenuItem    = "Pin this Post"
	UnpinPostMenuItem  = "Unpin this Post"
	ReportPostMenuItem = "Report"
)

// Exposed Method to get Post Menu for Owner who are CMs also
func GetIsOwnerIsCmPostMenuItems(is_pinned bool) []string {
	menuItems := []string{DeletePostMenuItem}

	if is_pinned {
		menuItems = append(menuItems, UnpinPostMenuItem)
	} else {
		menuItems = append(menuItems, PinPostMenuItem)
	}
	return menuItems
}

// Exposed Method to get Post Menu for Owner who is not a CM
func GetIsOwnerNotIsCmPostMenuItems() []string {
	return []string{DeletePostMenuItem}
}

// Exposed Method to get Post Menu for CMs who are not owners
func GetNotIsOwnerIsCmPostMenuItems(is_pinned bool) []string {
	menuItems := []string{}

	if is_pinned {
		menuItems = append(menuItems, UnpinPostMenuItem)
	} else {
		menuItems = append(menuItems, PinPostMenuItem)
	}

	menuItems = append(menuItems, DeletePostMenuItem)
	return menuItems
}

// Exposed Method to get Post Menu for members
func GetNotIsOwnerNotIsCmPostMenuItems() []string {
	return []string{ReportPostMenuItem}
}
