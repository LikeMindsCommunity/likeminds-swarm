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

func GetIsOwnerIsCmPostMenuItems() []string {
	return []string{DeletePostMenuItem, PinPostMenuItem}
}

func GetIsOwnerNotIsCmPostMenuItems() []string {
	return []string{DeletePostMenuItem}
}

func GetNotIsOwnerIsCmPostMenuItems() []string {
	return []string{PinPostMenuItem, DeletePostMenuItem}
}

func GetNotIsOwnerNotIsCmPostMenuItems() []string {
	return []string{ReportPostMenuItem}
}
