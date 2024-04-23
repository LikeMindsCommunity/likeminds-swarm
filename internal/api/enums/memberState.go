package enums

const (
	CM     int = 1
	MEMBER int = 4
	GUEST  int = 0
)

// checks whether the member state is CM or not
func IsCM(state int) bool {
	return state == CM
}
