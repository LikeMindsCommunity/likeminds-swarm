package constants

const CommentBaseLevel int = 0
const CommentAllowedLevel int = 1
const CommentAllowedErrorMessage string = "Only one level of replies are allowed"
const CommentEntityType string = "comment"

const (
	DeleteCommentMenuItem = "Delete"
	ReportCommentMenuItem = "Report"
)

// Exposed Method to get Comment Menu for owner
func GetIsOwnerCommentMenuItems() []string {
	return []string{DeleteCommentMenuItem}
}

// Exposed Method to get Comment Menu for CM who is not owner
func GetNotIsOwnerIsCmCommentMenuItems() []string {
	return []string{DeleteCommentMenuItem}
}

// Exposed Method to get Comment Menu for members
func GetNotIsOwnerNotIsCmCommentMenuItems() []string {
	return []string{ReportCommentMenuItem}
}
