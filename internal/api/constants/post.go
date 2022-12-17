package constants

const PostEntityType string = "post"
const (
	ImageWidget int = iota + 1
	VideoWidget
	DocumentWidget
)

const (
	DeletePostMenuItem = "Delete Post"
	PinPostMenuItem    = "Pin this Post"
	UnpinPostMenuItem  = "Unpin this Post"
	ReportPostMenuItem = "Report"
)

func GetIsOwnerIsCmPostMenuItems(is_pinned bool) []string {
	menuItems := []string{DeletePostMenuItem}

	if is_pinned {
		menuItems = append(menuItems, UnpinPostMenuItem)
	} else {
		menuItems = append(menuItems, PinPostMenuItem)
	}
	return menuItems
}

func GetIsOwnerNotIsCmPostMenuItems() []string {
	return []string{DeletePostMenuItem}
}

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

func GetNotIsOwnerNotIsCmPostMenuItems() []string {
	return []string{ReportPostMenuItem}
}
