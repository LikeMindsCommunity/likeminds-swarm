package searchElastic

// Interface for ElasticSearch Helper
type EsHelper interface {
	CreateIndex(index string) error
	DeleteIndex(index string) error
	IndexDocument(document interface{}, documentId string, index string) error
	InsertManyDocuments(document map[string]interface{}, index string) error
	DeleteDocument(documentId string, index string) error
	ExecuteQuery(query string, index string) map[string]interface{}
	UpdateByQuery(query string, index string) error
	DeleteByQuery(query string, index string) error
}
