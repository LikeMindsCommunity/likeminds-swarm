package searchElastic

import "context"

// Interface for Like Repository
type EsHelper interface {
	CreateIndex(index string) error
	DeleteIndex(index string) error
	InsertDocument(ctx context.Context, document interface{}, documentId string, index string) error
	InsertManyDocuments(document map[string]interface{}, index string) error
	UpdateDocument(ctx context.Context, document interface{}, documentId string, index string) error
	UpdateByQuery(query string, index string) error
	IndexDocument(ctx context.Context, document interface{}, documentId string, index string) error
	DeleteDocument(ctx context.Context, documentId string, index string) error
	ExecuteQuery(query string, index string) map[string]interface{}
	DeleteByQuery(query string, index string) error
}
