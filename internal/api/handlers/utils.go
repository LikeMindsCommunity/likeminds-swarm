package handlers

import (
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

func fetchPaginationParams(c *gin.Context) (int, int, error) {
	// fetch and validate query params
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		return 0, 0, err
	}

	page_size, err := strconv.Atoi(c.DefaultQuery("page_size", "0"))
	if err != nil {
		return 0, 0, err
	}

	if page <= 0 || page_size <= 0 {
		return 1, 10, nil
	}

	return page, page_size, nil
}

func generatePageFilterOptions(c *gin.Context) (map[string]interface{}, error) {
	// fetch pagination query params
	page, page_size, err := fetchPaginationParams(c)
	if err != nil {
		return nil, err
	}

	// page filter options
	filter_options := gin.H{
		"$sort": gin.H{
			"created_at": -1,
		},
		"$skip":  page_size * (page - 1),
		"$limit": page_size,
	}

	return filter_options, nil
}

func getTaggedUsers(text string) ([]string, error) {
	tagged_members := []string{}
	pattern, err := regexp.Compile("route://[member member_profile]+/(?P<user_id>[0-9]+)")
	if err != nil {
		return nil, err
	}

	allSubstringMatches := pattern.FindAllStringSubmatch(text, -1)

	for _, occurance := range allSubstringMatches {
		tagged_members = append(tagged_members, occurance[1])
	}

	return tagged_members, nil
}
