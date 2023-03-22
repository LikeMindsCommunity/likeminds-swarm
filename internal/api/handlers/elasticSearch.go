package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
)

func ParsePostIndexData(Post *entities.Post) searchElastic.PostIndex {
	return searchElastic.PostIndex{
		Id:          Post.ID.Hex(),
		Text:        Post.Text,
		Heading:     Post.Heading,
		ChatroomId:  Post.ChatroomId,
		CommunityId: Post.CommunityId,
		IsPinned:    Post.IsPinned,
		UserId:      Post.UserId,
		Attachments: Post.Attachments,
		CreatedAt:   Post.CreatedAt,
		UpdatedAt:   Post.UpdatedAt,
	}
}

// Exposed method to create post search query
func GetPostFilterQuery(page int, page_size int, search_type string, search string, chatroom_ids string, community_id int) string {
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
								"query": "%s"
							}
						}
					}
				]
			}
		},`, search_type, search)
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
	`, from, page_size, communityQuery, searchQuery, chatroomQuery)
}

// Exposed method to create post search query
func GetSelfPostFilterQuery(page int, page_size int, search_type string, search string, member_id string, community_id int) string {
	from := page_size * (page - 1)
	searchQuery := ""
	if search != "" && search_type != "" {
		searchQuery = fmt.Sprintf(`{
			"match": {
				"%s": {
					"query": "%s"
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
