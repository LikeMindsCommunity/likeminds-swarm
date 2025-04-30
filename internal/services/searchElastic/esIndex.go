package searchElastic

import (
	"time"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Struct for Elasticsearch Post Index fields
type PostIndex struct {
	Id                 string               `json:"id"`
	Text               string               `json:"text"`
	Heading            string               `json:"heading"`
	TopicIds           []primitive.ObjectID `json:"topic_ids"`
	ChatroomId         int                  `json:"chatroom_id"`
	CommunityId        int                  `json:"community_id"`
	IsPinned           bool                 `json:"is_pinned"`
	IsRepost           bool                 `json:"is_repost"`
	IsAnonymous        bool                 `json:"is_anonymous"`
	IsHidden           bool                 `json:"is_hidden"`
	UserId             string               `json:"user_id"`
	OriginalAuthorUUID string               `json:"original_author_uuid,omitempty"`
	Attachments        interface{}          `json:"attachments"`
	PostShareCount     int                  `json:"post_share_count"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// Struct for Elasticsearch Topic Index fields
type TopicIndex struct {
	Id              string    `json:"id"`
	Name            string    `json:"name"`
	IsEnabled       bool      `json:"is_enabled"`
	Priority        float32   `json:"priority"`
	IsSearchable    bool      `json:"is_searchable"`
	ParentId        string    `json:"parent_id"`
	ParentName      string    `json:"parent_name"`
	AllParentIds    []string  `json:"all_parent_ids"`
	Level           int       `json:"level"`
	WidgetId        string    `json:"widget_id"`
	TotalChildCount int       `json:"total_child_count"`
	Access          string    `json:"access"`
	CommunityId     int       `json:"community_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	NumberOfPosts   int       `json:"number_of_posts"`
}

// Struct for Elasticsearch Widget Index fields
type WidgetIndex struct {
	Id               string      `json:"id"`
	CreatedByLM      bool        `json:"created_by_lm"`
	ParentEntityID   string      `json:"parent_entity_id"`
	ParentEntityType string      `json:"parent_entity_type"`
	MetaData         interface{} `json:"metadata"`
	CommunityId      int         `json:"community_id"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
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
	constants.TopicIndexName: `
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
				"name": {
					"type": "text",
					"analyzer": "autocomplete",
					"fields": {
						"raw": {
							"type": "keyword"
						}
					}
				}
			}
		}
	}
	`,
	constants.WidgetIndexName: `
	{
		"mappings": {
			"properties": {
				"metadata": {
					"type": "flat_object"
				}
			}
		}
	}
	`,
}
