package constants

// Post Menu Items
const (
	DeletePostMenuItemName = "DeletePost"
	PinPostMenuItemName    = "PinPost"
	UnpinPostMenuItemName  = "UnpinPost"
	ReportPostMenuItemName = "ReportPost"
	EditPostMenuItemName   = "EditPost"
	BlockUserMenuItemName  = "BlockUser"
)

const (
	DeletePostMenuItemTitle = "Delete %s"
	PinPostMenuItemTitle    = "Pin This %s"
	UnpinPostMenuItemTitle  = "Unpin This %s"
	ReportPostMenuItemTitle = "Report"
	EditPostMenuItemTitle   = "Edit %s"
	BlockUserMenuItemTitle  = "Block"
)

// Comment Menu Items
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

// Pending Post Menu Items
const (
	EditPendingPostMenuItemName   = "EditPendingPost"
	DeletePendingPostMenuItemName = "DeletePendingPost"
)

const (
	EditPendingPostMenuItemTitle   = "Edit"
	DeletePendingPostMenuItemTitle = "Delete"
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
	EditPendingPostMenuItemId
	DeletePendingPostMenuItemId
	BlockUserMenuItemId
)
