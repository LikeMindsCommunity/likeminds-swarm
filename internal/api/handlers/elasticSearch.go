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
		CommunityId: fmt.Sprint(Post.CommunityId),
		IsPinned:    Post.IsPinned,
		UserId:      Post.UserId,
		Attachments: Post.Attachments,
		CreatedAt:   Post.CreatedAt,
		UpdatedAt:   Post.UpdatedAt,
	}
}
