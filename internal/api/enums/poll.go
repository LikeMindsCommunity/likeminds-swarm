package enums

// constants for poll_type
const (
	InstantPollType  string = "instant"
	DeferredPollType string = "deferred"
)

// function to check if poll type is valid
func IsPollTypeValid(pollType string) bool {
	switch pollType {
	case InstantPollType, DeferredPollType:
		return true
	}
	return false
}

// constants for multiple_select_state in poll
const (
	ExactlySelectStateType string = "exactly"
	AtMaxSelectStateType   string = "at_max"
	AtLeastSelectStateType string = "at_least"
)

// function to check if poll multiple select state is valid
func IsPollMultipleSelectStateValid(pollType string) bool {
	switch pollType {
	case ExactlySelectStateType, AtLeastSelectStateType,
		AtMaxSelectStateType:
		return true
	}
	return false
}
