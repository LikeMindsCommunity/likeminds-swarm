package searchElastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func constructQuery(query string) *strings.Reader {
	// Check for JSON errors
	isValid := json.Valid([]byte(query)) // returns bool

	// Default query is "{}" if JSON is invalid
	if !isValid {
		fmt.Println("Search(Elastic): constructQuery() ERROR: query string not valid:", query)
		fmt.Println("Search(Elastic): Using default match_all query")
		query = "{}"
	}

	// Build a new string from JSON query
	var b strings.Builder
	b.WriteString(query)

	// Instantiate a *strings.Reader object from string
	read := strings.NewReader(b.String())

	// Return a *strings.Reader object
	return read
}

// Exposed method to execute a query in ElasticSearch
func (esHelper *esHelper) ExecuteQuery(query string, index string) map[string]interface{} {
	// Create a context object for the API calls
	ctx := context.Background()

	// Pass the query string to the function and have it return a Reader object
	read := constructQuery(query)

	// Instantiate a map interface object for storing returned documents
	var mapResp map[string]interface{}

	var buf bytes.Buffer

	// Attempt to encode the JSON query and look for errors
	if err := json.NewEncoder(&buf).Encode(&read); err != nil {
		log.Error(fmt.Sprintf("Search(Elastic): json.NewEncoder() ERROR: %v", err))
	} else {
		client := esHelper.esClient

		// Pass the JSON query to the Golang client's Search() method
		res, err := client.Search(
			client.Search.WithContext(ctx),
			client.Search.WithIndex(index),
			client.Search.WithBody(read),
			client.Search.WithTrackTotalHits(true),
			client.Search.WithPretty(),
		)

		// Check for any errors returned by API call to Elasticsearch
		if err != nil {
			log.Error(fmt.Sprintf("Search(Elastic): Elasticsearch Search() API ERROR: %v", err))

			// If no errors are returned, parse esapi.Response object
		} else {
			// Close the result body when the function call is complete
			defer res.Body.Close()

			// Decode the JSON response and using a pointer
			json.NewDecoder(res.Body).Decode(&mapResp)
		}
	}

	return mapResp
}

// Exposed method to create a new index in ElasticSearch
func (esHelper *esHelper) CreateIndex(index string) error {
	res, err := esHelper.esClient.Indices.Exists([]string{index})
	if err != nil {
		return fmt.Errorf("Search(Elastic): cannot check index existence: %w", err)
	}
	if res.StatusCode == 200 {
		return nil
	}

	if res.StatusCode != 404 {
		return fmt.Errorf("Search(Elastic): error in index existence response: %s", res.String())
	}

	if mapping, ok := IndexMapping[index]; ok {
		res, err = esHelper.esClient.Indices.Create(index, esHelper.esClient.Indices.Create.WithBody(strings.NewReader(mapping)))
	} else {
		res, err = esHelper.esClient.Indices.Create(index)
	}

	if err != nil {
		return fmt.Errorf("Search(Elastic): cannot create index: %w", err)
	}
	if res.IsError() {
		return fmt.Errorf("Search(Elastic): error in index creation response: %s", res.String())
	}

	return nil
}

// Exposed method to delete an existing index in ElasticSearch
func (esHelper *esHelper) DeleteIndex(index string) error {
	_, err := esHelper.esClient.Indices.Delete([]string{index})
	if err != nil {
		return fmt.Errorf("Search(Elastic): cannot delete index: %w", err)
	}

	return nil
}

// Exposed method to insert a new document in ElasticSearch
func (esHelper *esHelper) InsertDocument(document interface{}, documentId string, index string) error {

	err := esHelper.CreateIndex(index)
	if err != nil {
		return err
	}

	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("Search(Elastic): insert: marshall: %w", err)
	}

	req := esapi.CreateRequest{
		Index:      index,
		DocumentID: documentId,
		Body:       bytes.NewReader(body),
	}

	ctx, cancel := context.WithTimeout(context.Background(), esHelper.timeout)
	defer cancel()

	res, err := req.Do(ctx, esHelper.esClient)
	if err != nil {
		return fmt.Errorf("Search(Elastic): insert: request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Search(Elastic): insert: response: %s", res.String())
	}

	return nil
}

// Exposed method to update an existing document in ElasticSearch
func (esHelper *esHelper) UpdateDocument(ctx context.Context, document interface{}, documentId string, index string) error {

	err := esHelper.CreateIndex(index)
	if err != nil {
		return err
	}

	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("Search(Elastic): update: marshall: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: documentId,
		Body:       bytes.NewReader([]byte(fmt.Sprintf(`{"doc":%s}`, body))),
	}

	ctx, cancel := context.WithTimeout(ctx, esHelper.timeout)
	defer cancel()

	res, err := req.Do(ctx, esHelper.esClient)
	if err != nil {
		return fmt.Errorf("Search(Elastic): update: request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Search(Elastic): update: response: %s", res.String())
	}

	return nil
}

// Exposed method to index a document in ElasticSearch (create if not exists, update if exists)
func (esHelper *esHelper) IndexDocument(document interface{}, documentId string, index string) error {

	err := esHelper.CreateIndex(index)
	if err != nil {
		return err
	}

	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("Search(Elastic): index: marshall: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentId,
		Body:       bytes.NewReader(body),
	}

	ctx, cancel := context.WithTimeout(context.Background(), esHelper.timeout)
	defer cancel()

	res, err := req.Do(ctx, esHelper.esClient)
	if err != nil {
		return fmt.Errorf("Search(Elastic): index: request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Search(Elastic): index: response: %s", res.String())
	}

	return nil
}

// Exposed method to delete an existing document in ElasticSearch
func (esHelper *esHelper) DeleteDocument(ctx context.Context, documentId string, index string) error {

	err := esHelper.CreateIndex(index)
	if err != nil {
		return err
	}

	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: documentId,
	}

	ctx, cancel := context.WithTimeout(ctx, esHelper.timeout)
	defer cancel()

	res, err := req.Do(ctx, esHelper.esClient)
	if err != nil {
		return fmt.Errorf("Search(Elastic): delete: request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Search(Elastic): delete: response: %s", res.String())
	}

	return nil
}

// Structure for ElasticSearch Helper
type esHelper struct {
	esClient *elasticsearch.Client
	timeout  time.Duration
}

// Exposed Method to Create New ElasticSearch Helper
func NewESHelper(es *elasticsearch.Client) EsHelper {
	return &esHelper{
		esClient: es,
		timeout:  time.Second * 10,
	}
}
