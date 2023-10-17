package constants

const (
	DeletePostMenuItemName = "DeletePost"
	PinPostMenuItemName    = "PinPost"
	UnpinPostMenuItemName  = "UnpinPost"
	ReportPostMenuItemName = "ReportPost"
	EditPostMenuItemName   = "EditPost"
)

const (
	DeletePostMenuItemTitle = "Delete %s"
	PinPostMenuItemTitle    = "Pin This %s"
	UnpinPostMenuItemTitle  = "Unpin This %s"
	ReportPostMenuItemTitle = "Report"
	EditPostMenuItemTitle   = "Edit %s"
)

const (
	DeleteCommentMenuItemName = "DeleteComment"
	ReportCommentMenuItemName = "ReportComment"
	EditCommentMenuItemName   = "EditComment"
)

const (
	DeleteCommentMenuItemTitle = "Delete"
	ReportCommentMenuItemTitle = "Report"
	EditCommentMenuItemTitle   = "Edit"
)

const (
	DeletePostMenuItemId int = iota + 1
	PinPostMenuItemId
	UnpinPostMenuItemId
	ReportPostMenuItemId
	EditPostMenuItemId
	DeleteCommentMenuItemId
	ReportCommentMenuItemId
	EditCommentMenuItemId
)
