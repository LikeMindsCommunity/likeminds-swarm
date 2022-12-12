package constants

const CommentBaseLevel int = 0
const CommentAllowedLevel int = 1
const CommentAllowedErrorMessage string = "Only one level of replies are allowed"
const CommentEntityType string = "comment"

const (
	DeleteCommentMenuItem = "Delete"
	ReportCommentMenuItem = "Report"
)

func GetIsOwnerCommentMenuItems() []string {
	return []string{DeleteCommentMenuItem}
}

func GetNotIsOwnerIsCmCommentMenuItems() []string {
	return []string{DeleteCommentMenuItem}
}

func GetNotIsOwnerNotIsCmCommentMenuItems() []string {
	return []string{ReportCommentMenuItem}
}
