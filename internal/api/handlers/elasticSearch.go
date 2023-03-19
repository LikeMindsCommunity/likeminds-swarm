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
func GetPostFilterQuery(page int, page_size int, search_type string, search string, chatroom_ids []int) string {
	from := page_size * (page - 1)

	return fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"sort": {
			"_score": {
				"order": "desc"
			},
			"updated_at": {
				"order": "desc"
			}
		},
		"query": {
			"bool": {
				"must": [
					{
						"bool": {
							"should": [
								{
									"match": {
										%s: {
											"query": %s,
											"analyzer": "standard"
										}
									}
								}
							]
						}
					}
				],
				"must_not": [
					{
						"term": {chatroom_id: %v}
					}
				]
			}
		}
	}
	`, from, page_size, search_type, search, chatroom_ids)
}
