package handlers

import (
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
)

type FeedHandlers struct {
	likeHelper     interfaces.LikeHelper
	commentHelper  interfaces.CommentHelper
	postHelper     interfaces.PostHelper
	activityHelper interfaces.ActivityHelper
	saveHelper     interfaces.SaveHelper
}

func NewFeedHandlers(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	postHelper interfaces.PostHelper, saveHelper interfaces.SaveHelper, activityHelper interfaces.ActivityHelper) *FeedHandlers {
	return &FeedHandlers{
		likeHelper:     likeHelper,
		commentHelper:  commentHelper,
		postHelper:     postHelper,
		saveHelper:     saveHelper,
		activityHelper: activityHelper,
	}
}

func fetchPaginationParams(c *gin.Context) (int, int, error) {
	// fetch and validate query params
	ParamPage := c.DefaultQuery("page", "0")
	ParamPageSize := c.DefaultQuery("page_size", "0")
	if ParamPage == "" {
		ParamPage = "0"
	}

	if ParamPageSize == "" {
		ParamPageSize = "0"
	}

	page, err := strconv.Atoi(ParamPage)
	if err != nil {
		return 0, 0, err
	}

	page_size, err := strconv.Atoi(ParamPageSize)
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

func parseMenuItems(menu_items []string) []requests.MenuResponse {
	output_menu_items := []requests.MenuResponse{}

	for _, value := range menu_items {
		menu_item := requests.MenuResponse{}
		menu_item.Title = value
		output_menu_items = append(output_menu_items, menu_item)
	}

	return output_menu_items
}

func getEntityMenuItems(entity_type string, is_cm bool, is_owner bool, is_pinned bool) []string {
	var output_menu_items []string
	switch entity_type {
	case constants.PostEntityType:
		if is_owner && is_cm {
			output_menu_items = constants.GetIsOwnerIsCmPostMenuItems()
		}

		if is_owner && !is_cm {
			output_menu_items = constants.GetIsOwnerNotIsCmPostMenuItems()
		}

		if !is_owner && is_cm {
			output_menu_items = constants.GetNotIsOwnerIsCmPostMenuItems()
		}

		if !is_owner && !is_cm {
			output_menu_items = constants.GetNotIsOwnerNotIsCmPostMenuItems()
		}

	case constants.CommentEntityType:
		if is_owner {
			output_menu_items = constants.GetIsOwnerCommentMenuItems()
		}

		if !is_owner && is_cm {
			output_menu_items = constants.GetNotIsOwnerIsCmCommentMenuItems()
		}

		if !is_owner && !is_cm {
			output_menu_items = constants.GetNotIsOwnerNotIsCmCommentMenuItems()
		}
	}

	return output_menu_items
}

func checkIfFibonacciNumber(num int) bool {
	var n3, n1, n2 int = 0, 0, 1

	if num == n1 || num == n2 {
		return true
	}

	n3 = n1 + n2

	for {
		if n3 > num {
			break
		}

		if n3 == num {
			return true
		}
		n1 = n2
		n2 = n3
		n3 = n1 + n2
	}

	return false
}
