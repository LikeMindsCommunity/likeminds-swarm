package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Internal Method to get TLS Configuration for Database Connection
func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := os.ReadFile(caFile)
	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)
	if !ok {
		return tlsConfig, errors.New("failed parsing pem file")
	}

	return tlsConfig, nil
}

// Exposed Method to Initiate DB Connection
func InitiateDB() *mongo.Database {
	var dbInstance *mongo.Database = nil

	if dbInstance == nil {
		connectionURI := environment.GoDotEnvVariable("MONGODB_URI")
		if connectionURI == "" {
			log.Fatal("Database(Mongo): You must set your 'MONGODB_URI' environment variable.")
		}

		db_name := environment.GoDotEnvVariable("DB_NAME")
		if db_name == "" {
			log.Fatal("Database(Mongo): You must set your 'DB_NAME' environment variable.")
		}

		var client *mongo.Client
		var err error

		cloud_provider := environment.GoDotEnvVariable("CLOUD_PROVIDER")
		if cloud_provider == "" || cloud_provider == "AWS" {
			caFilePath := environment.GoDotEnvVariable("CA_FILE_PATH")
			if caFilePath == "" {
				log.Fatal("Database(Mongo): You must set your 'CA_FILE_PATH' environment variable.")
			}

			tlsConfig, err := getCustomTLSConfig(caFilePath)
			if err != nil {
				log.Fatal(fmt.Sprintf("Database(Mongo): Failed getting TLS configuration: %v", err))
			}

			// Create a new client and connect to the server
			client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionURI).SetTLSConfig(tlsConfig))
			if err != nil {
				log.Panic(err)
				panic(err)
			}
		} else {
			// Create a new client and connect to the server
			client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionURI))
			if err != nil {
				log.Panic(err)
				panic(err)
			}
		}

		// Ping the primary
		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			log.Panic(err)
			panic(err)
		}
		log.Info("Database(Mongo): Successfully connected and pinged.")

		dbInstance = client.Database(db_name)
	}

	return dbInstance
}
