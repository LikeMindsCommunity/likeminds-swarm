package enums

const (
	ADMIN  int = 1
	MEMBER int = 4
	GUEST  int = 0
)

// checks whether the member state is ADMIN or not
func IsAdmin(state int) bool {
	return state == ADMIN
}
