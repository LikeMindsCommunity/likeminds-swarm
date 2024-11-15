package utils

// Error Messages
const (
	NotAuthorizedError         = "You are not authorized to perform this operation."
	InvalidRequestError        = "Invalid request."
	InvalidPostIDError         = "Invalid post_id sent."
	InvalidCommentIDError      = "Invalid comment_id sent."
	NsfwContentInImageError    = "This post could not be submitted as %simage/s seems to contain NSFW content."
	ErrorGuestAccessNotAllowed = "Guest access is not allowed."
	PendingPostCreationError   = "Some error occurred in creation of pending post."
	PendingPostUpdationError   = "Some error occurred in updation of pending post."
	BlockedUserTagError        = "You cannot tag a blocked user!"
	BlockingUserTagError       = "You cannot tag a user who blocked you!"
	PostHiddenCannotPinError   = "This post cannot be pinned because it is currently hidden. Please unhide the post."
	PostIsHiddenError          = "The post is hidden."
	ErrorSomethingWentWrong    = "Something went wrong."
)
