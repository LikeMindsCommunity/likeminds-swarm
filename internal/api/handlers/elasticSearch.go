package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ParsePostIndexData(Post *entities.Post) searchElastic.PostIndex {
	postEntity := searchElastic.PostIndex{
		Id:             Post.ID.Hex(),
		Text:           Post.Text,
		Heading:        Post.Heading,
		TopicIds:       Post.TopicIds,
		ChatroomId:     Post.ChatroomId,
		CommunityId:    Post.CommunityId,
		IsPinned:       Post.IsPinned,
		IsRepost:       Post.IsRepost,
		IsAnonymous:    Post.IsAnonymous,
		IsHidden:       Post.IsHidden,
		UserId:         Post.UserId,
		Attachments:    Post.Attachments,
		PostShareCount: Post.PostShareCount,
		CreatedAt:      Post.CreatedAt,
		UpdatedAt:      Post.UpdatedAt,
	}

	if Post.OriginalAuthorUUID != "" {
		postEntity.OriginalAuthorUUID = Post.OriginalAuthorUUID
	}

	return postEntity
}

// Exposed method to create post search query
func GetPostFilterQuery(userId string, page int, page_size int, search_type string, search string, chatroom_ids string,
	community_id int, excludedUserIds []string, isCm bool,
) string {

	from := page_size * (page - 1)

	chatroomQuery := ""
	if chatroom_ids != "" {
		chatroomQuery = fmt.Sprintf(`{
			"script" : {
				"script" : {
				"inline" : "for(int chatroom_id: doc['chatroom_id']) { boolean matches = true; for(int excluded_chatroom_id: params.chatroom_ids){if(chatroom_id==excluded_chatroom_id) matches = false;} if(matches) return true;} ",
				"lang"   : "painless",
				"params" : {"chatroom_ids": %s}
				}
			}
		}`, chatroom_ids)
	}

	communityQuery := ""
	if community_id != 0 {
		communityQuery = fmt.Sprintf(`{
			"bool": {
				"must": [
					{
						"match": {
							"community_id": {
								"query": %d
							}
						}
					}
				]
			}
		},`, community_id)
	}

	searchQuery := ""
	if search != "" && search_type != "" {
		searchQuery = fmt.Sprintf(`{
			"bool": {
				"must": [
					{
						"match": {
							"%s": {
								"query": "%s",
								"analyzer": "standard"
							}
						}
					}
				]
			}
		},`, search_type, search)
	}

	excludedUserPostsQuery := ""
	if len(excludedUserIds) > 0 {
		excludedUserPostsQuery = fmt.Sprintf(`
		{
			"terms": {
				"user_id.keyword": %s
			}
		}`, utils.ParseStringArrayToString(excludedUserIds))
	}

	hiddenPostsQuery := ""
	if !isCm {
		hiddenPostsQuery = fmt.Sprintf(`,
		"should": [
			{ "term": { "is_hidden": false}},
			{
				"bool": {
					"must": [
						{ "term": { "is_hidden": true} },
						{ "term": { "user_id.keyword": "%s" }}
					]
				}
			}
		],
		"minimum_should_match": 1`, userId)
	}

	return fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"sort": [
			{"updated_at": {"order": "desc"}}
		],
		"query": {
			"bool": {
				"must_not": [
					%s
				],
				"must": [
					%s
					%s
					%s
				]
					%s
			}
		}
	}`, from, page_size, excludedUserPostsQuery, communityQuery, searchQuery, chatroomQuery, hiddenPostsQuery)
}

// Exposed method to create post search query
func GetSelfPostFilterQuery(page int, page_size int, search_type string, search string, member_id string, community_id int) string {
	from := page_size * (page - 1)
	searchQuery := ""
	if search != "" && search_type != "" {
		searchQuery = fmt.Sprintf(`{
			"match": {
				"%s": {
					"query": "%s",
					"analyzer": "standard"
				}
			}
		},`, search_type, search)
	}

	communityQuery := ""
	if community_id != 0 {
		communityQuery = fmt.Sprintf(`{
			"bool": {
				"must": [
					{
						"match": {
							"community_id": {
								"query": %d
							}
						}
					}
				]
			}
		},`, community_id)
	}

	userQuery := ""
	if member_id != "" {
		userQuery = fmt.Sprintf(`{
			"match": {
				"user_id": {
					"query": "%s"
				}
			}
		}`, member_id)
	}

	return fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"sort": [
			{"updated_at": {"order": "desc"}}
		],
		"query": {
			"bool": {
				"must": [
					%s
					%s
					%s
				]
			}
		}
	}
	`, from, page_size, communityQuery, searchQuery, userQuery)
}

func ParseTopicIndexData(postHelper interfaces.PostHelper, Topic *entities.Topic, updatePostCount bool) searchElastic.TopicIndex {

	topicIndex := searchElastic.TopicIndex{
		Id:              Topic.ID.Hex(),
		Name:            Topic.Name,
		IsEnabled:       Topic.IsEnabled,
		CommunityId:     Topic.CommunityId,
		Priority:        Topic.Priority,
		IsSearchable:    Topic.IsSearchable,
		ParentName:      Topic.ParentName,
		Level:           Topic.Level,
		TotalChildCount: Topic.TotalChildCount,
		Access:          Topic.Access,
		CreatedAt:       Topic.CreatedAt,
		UpdatedAt:       Topic.UpdatedAt,
	}

	if Topic.ParentId != primitive.NilObjectID {
		topicIndex.ParentId = Topic.ParentId.Hex()
	}

	if Topic.AllParentIds != nil {
		topicIndex.AllParentIds = helpers.ParseObjectIdsToStringArray(Topic.AllParentIds)
	}

	if Topic.WidgetId != primitive.NilObjectID {
		topicIndex.WidgetId = Topic.WidgetId.Hex()
	}

	// if updatePostCount is true, then fetch the posts with topic id and update the count
	if updatePostCount {

		postResults, err := fetchPostsWithTopicID(postHelper, Topic.ID, Topic.CommunityId)
		if err != nil {
			logging.Error(fmt.Sprint("Error while fetching posts with topic id: ", err.Error()))
		}
		topicIndex.NumberOfPosts = len(postResults)
	}

	return topicIndex
}

// Exposed method to fetch topics with topic id from elastic search
func GetTopicIdsFilterQuery(topicIds []string, communityId int) string {
	return fmt.Sprintf(`
	{
		"sort": [
			{"updated_at": {"order": "desc"}}
		],
		"query": {
			"bool": {
				"must": [
					{
						"match": {"community_id": {"query": %d}}
					},
					{
						"terms": {
							"id" : %s
						}
					}
				]
			}
		}
	}`, communityId, utils.ParseStringArrayToString(topicIds))
}

// Exposed method to create topic search query.
// Note : Everytime a new field is added in mongodb, it doesn't cause any problem there but it sometimes causes problem in elastic search because in older documents the newly added field is null. So take care of this
func GetTopicFilterQuery(page int, pageSize int, searchType string, search string, communityId int, filterIsEnabled bool,
	isEnabled bool, minPosts int, orderByParams []string, parentTopicId string, isCM bool, memberRole string) string {

	from := pageSize * (page - 1)

	communityQuery := ""
	if communityId != 0 {
		communityQuery = fmt.Sprintf(`{
			"match": {
				"community_id": {
					"query": %d
				}
			}
		}`, communityId)
	}

	isEnabledQuery := ""
	if filterIsEnabled {
		isEnabledQuery = fmt.Sprintf(`,{
			"match": {
				"is_enabled": {
					"query": %t
				}
			}
		}`, isEnabled)
	}

	searchQuery := ""
	if search != "" && searchType != "" {
		searchQuery = fmt.Sprintf(`,
		{ 
			"match": { "%s": { "query": "%s", "analyzer": "standard" } }
		},
		{ 
			"match": { "is_searchable": { "query": true } }
		}`, searchType, search)
	}

	minPostsQuery := fmt.Sprintf(`, {
		"range": {
			"number_of_posts": {
				"gte": %d
			}
		}
	}`, minPosts)

	sortQuery := getSortQueryFromOrderByParams(orderByParams)

	parentTopicQuery := ""
	if parentTopicId != "" {
		parentTopicQuery = fmt.Sprintf(`,{
			"match": {
				"parent_id": {
					"query": "%s"
				}
			}
		}`, parentTopicId)
	}

	levelQuery := ""
	if parentTopicQuery == "" && searchQuery == "" {
		levelQuery = `,{
			"match": {
				"level": {
					"query": 0
				}
			}
		}`
	}

	var accessQuery string
	var accessValues []string

	switch memberRole {
	case utils.CMRole:
		accessValues = []string{enums.ONLY_CM_TOPIC_ACCESS, enums.EVERYONE_TOPIC_ACCESS, ""}
	case utils.MemberRole:
		accessValues = []string{enums.EVERYONE_TOPIC_ACCESS, ""}
	default:
		accessValues = []string{enums.EVERYONE_TOPIC_ACCESS, ""}
	}

	accessStringArray := utils.ParseStringArrayToString(accessValues)
	accessQuery = fmt.Sprintf(`,
		{
		  "bool": {
		    "should": [
		      {
		        "terms": {
		          "access.keyword": %s
		        }
		      },
		      {
		        "bool": {
		          "must_not": {
		            "exists": {
		              "field": "access"
		            }
		          }
		        }
		      }
		    ]
		  }
		}`, accessStringArray)

	return fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"sort": [%s],
		"query": {
			"bool": {
				"must": [
					%s
					%s
					%s
					%s
					%s
					%s
					%s
				]
			}
		}
	}`, from, pageSize, sortQuery, communityQuery, isEnabledQuery, searchQuery, minPostsQuery, parentTopicQuery, levelQuery, accessQuery)
}

func getSortQueryFromOrderByParams(orderByParams []string) string {

	orderByQuery := ""
	for _, orderByParam := range orderByParams {

		switch orderByParam {
		case enums.OrderByAlphabeticalAsc:
			orderByQuery += `{"name.raw": {"order": "asc"}}`
		case enums.OrderByPriorityDesc:
			orderByQuery += `{"priority": {"order": "desc"}}`
		case enums.OrderByCreatedAtDesc:
			orderByQuery += `{"created_at": {"order": "desc"}}`
		case enums.OrderByNumberOfPostsDesc:
			orderByQuery += `{"number_of_posts": {"order": "desc"}}`
		}

		orderByQuery += ","
	}

	orderByQuery += `{"updated_at": {"order": "desc"}}`
	return orderByQuery
}

// Exposed method to get topics by their ids
func GetTopicsByIdQuery(topicIds string) string {
	return fmt.Sprintf(`
	{
		"query": {
			"terms": {
				"_id": %s
			}
		}
	}`, topicIds)
}

func ParseWidgetIndexData(Widget *entities.Widget) searchElastic.WidgetIndex {
	return searchElastic.WidgetIndex{
		Id:               Widget.ID.Hex(),
		CreatedByLM:      Widget.CreatedByLM,
		ParentEntityID:   Widget.ParentEntityID,
		ParentEntityType: Widget.ParentEntityType,
		MetaData:         Widget.MetaData,
		CommunityId:      Widget.CommunityId,
		CreatedAt:        Widget.CreatedAt,
		UpdatedAt:        Widget.UpdatedAt,
	}
}

// Exposed method to create widget search query
func GetWidgetFilterQuery(page int, pageSize int, communityId int, searchKey string, searchValue string) string {
	from := pageSize * (page - 1)

	communityQuery := ""
	if communityId != 0 {
		communityQuery = fmt.Sprintf(`{
			"match": {
				"community_id": {
					"query": %d
				}
			}
		}`, communityId)
	}

	searchQuery := ""
	if searchKey != "" && searchValue != "" {
		searchQuery = fmt.Sprintf(`,{
			"term": {
				%s: %s
			}
		}`, searchKey, searchValue)
	}

	return fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"sort": [
			{"updated_at": {"order": "desc"}}
		],
		"query": {
			"bool": {
				"must": [
					%s
					%s
				]
			}
		}
	}`, from, pageSize, communityQuery, searchQuery)
}

// Exposed query to fetch widgets by ids
func GetWidgetByIdsFilterQuery(communityId int, widgetIds string) string {

	searchQuery := fmt.Sprintf(`
	{
		"sort": [
			{"updated_at": {"order": "desc"}}
		],
		"query": {
			"bool": {
				"must": [
					{
						"match": {"community_id": {"query": %d}}
					},
					{
						"terms": {
							"id" : %s
						}
					}
				]
			}
		}
	}`, communityId, widgetIds)

	return searchQuery
}

// Exposed query to fetch widgets using parent entity id and type
func GetWidgetsByParentEntityFilterQuery(communityId int, parentEntityId string, parentEntityType string) string {

	searchQuery := fmt.Sprintf(`
	{
		"sort": [
			{"updated_at": {"order": "desc"}}
		],
		"query": {
			"bool": {
				"must": [
					{
						"match": {"community_id": {"query": %d}}
					},
					{
						"match": {"parent_entity_id": {"query": "%s"}}
					},
					{
						"match": {"parent_entity_type": {"query": "%s"}}
					}
				]
			}
		}
	}`, communityId, parentEntityId, parentEntityType)

	return searchQuery
}

// Exposed query to increment/decrement the count of posts in topics
func UpdatePostCountInTopicsQuery(topicIds string, increment bool) string {
	updatePostScript := ""
	if increment {
		updatePostScript = `
			"source": "ctx._source.number_of_posts += 1",
			"lang":   "painless"
		`
	} else {
		updatePostScript = `
			"source": "ctx._source.number_of_posts -= 1",
			"lang":   "painless"
		`
	}

	return fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": {
					"terms": {
						"_id": %s
					}
				}
			}
		},
		"script" : {
			%s
		}
	}`, topicIds, updatePostScript)
}
