package searchElastic

import (
	"time"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
)

// Struct for Elasticsearch Post Index fields
type PostIndex struct {
	Id          string      `json:"id"`
	Text        string      `json:"text"`
	Heading     string      `json:"heading"`
	ChatroomId  int         `json:"chatroom_id"`
	CommunityId int         `json:"community_id"`
	IsPinned    bool        `json:"is_pinned"`
	UserId      string      `json:"user_id"`
	Attachments interface{} `json:"attachments"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

var IndexMapping = map[string]string{
	constants.PostIndexName: `
	{
		"settings": {
			"analysis": {
				"analyzer": {
					"autocomplete": {
						"type": "custom",
						"tokenizer": "standard",
						"char_filter": [
							"html_strip"
						],
						"filter": [
							"lowercase",
							"edge_ngram_completion_filter"
						]
					}
				},
				"filter": {
					"edge_ngram_completion_filter": {
						"type": "edge_ngram",
						"min_gram": 1,
						"max_gram": 20
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"text": {
					"type": "text",
					"analyzer": "autocomplete"
				},
				"heading": {
					"type": "text",
					"analyzer": "autocomplete"
				}
			}
		}
	}
	`,
}
