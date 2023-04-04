package constants

const (
	DeletePostMenuItemName = "DeletePost"
	PinPostMenuItemName    = "PinPost"
	UnpinPostMenuItemName  = "UnpinPost"
	ReportPostMenuItemName = "ReportPost"
)

const (
	DeletePostMenuItemTitle = "Delete Post"
	PinPostMenuItemTitle    = "Pin this Post"
	UnpinPostMenuItemTitle  = "Unpin this Post"
	ReportPostMenuItemTitle = "Report"
)

const (
	DeleteCommentMenuItemName = "DeleteComment"
	ReportCommentMenuItemName = "ReportComment"
)

const (
	DeleteCommentMenuItemTitle = "Delete"
	ReportCommentMenuItemTitle = "Report"
)

const (
	DeletePostMenuItemId int = iota + 1
	PinPostMenuItemId
	UnpinPostMenuItemId
	ReportPostMenuItemId
	DeleteCommentMenuItemId
	ReportCommentMenuItemId
)
