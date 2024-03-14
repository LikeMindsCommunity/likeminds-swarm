package enums

// valid order_by for fetch topics request
const (
	OrderByAlphabeticalAsc   = "alphabetical_asc" // Default
	OrderByPriorityDesc      = "priority_desc"
	OrderByCreatedAtDesc     = "created_at_desc"
	OrderByNumberOfPostsDesc = "number_of_posts_desc"
)

// check if the order_by is valid
func IsValidOrderByParam(order_by string) bool {
	switch order_by {
	case OrderByAlphabeticalAsc, OrderByPriorityDesc, OrderByCreatedAtDesc, OrderByNumberOfPostsDesc:
		return true
	}
	return false
}
