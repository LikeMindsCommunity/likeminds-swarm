package searchElastic

import (
	"encoding/json"
	"log"

	// Import the Elasticsearch library packages
	"github.com/elastic/go-elasticsearch/v7"
)

func getESClusterInfo(client *elasticsearch.Client) {
	var (
		r map[string]interface{}
	)

	res, err := client.Info()
	if err != nil {
		log.Fatalf("Search(Elastic): Error getting response: %s", err)
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		log.Fatalf("Search(Elastic): Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Search(Elastic): Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	log.Printf("Search(Elastic): Client: %s", elasticsearch.Version)
	log.Printf("Search(Elastic): Server: %s", r["version"].(map[string]interface{})["number"])
}

func InitiateES() *elasticsearch.Client {
	var esInstance *elasticsearch.Client = nil

	if esInstance == nil {
		// Instantiate an Elasticsearch configuration
		cfg := elasticsearch.Config{
			Addresses: []string{
				"http://localhost:9200",
			},
		}

		// Instantiate a new Elasticsearch client object instance
		client, err := elasticsearch.NewClient(cfg)

		// Check for connection errors to the Elasticsearch cluster
		if err != nil {
			log.Fatalf("Search(Elastic): Elasticsearch connection error: %v", err)
		}

		// Get cluster info
		getESClusterInfo(client)

		esInstance = client
	}

	return esInstance
}
