package searchElastic

import "context"

// Interface for Like Repository
type EsHelper interface {
	CreateIndex(index string) error
	DeleteIndex(index string) error
	InsertDocument(ctx context.Context, document interface{}, documentId string, index string) error
	UpdateDocument(ctx context.Context, document interface{}, documentId string, index string) error
	DeleteDocument(ctx context.Context, documentId string, index string) error
	ExecuteQuery(query string, index string) map[string]interface{}
}
