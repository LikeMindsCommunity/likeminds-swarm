package constants

const (
	DeletePostMenuItemName = "DeletePost"
	PinPostMenuItemName    = "PinPost"
	UnpinPostMenuItemName  = "UnpinPost"
	ReportPostMenuItemName = "ReportPost"
	EditPostMenuItemName   = "EditPost"
)

const (
	DeletePostMenuItemTitle = "Delete Post"
	PinPostMenuItemTitle    = "Pin this Post"
	UnpinPostMenuItemTitle  = "Unpin this Post"
	ReportPostMenuItemTitle = "Report"
	EditPostMenuItemTitle   = "Edit Post"
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
