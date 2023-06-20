package environment

import (
	"os"
	"path/filepath"

	"log"

	"github.com/joho/godotenv"
)

// GoDotEnvVariable to load/read the .env file and return the value of the key
func GoDotEnvVariable(key string) string {
	// load .env file
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	environmentPath := filepath.Join(dir, "config/.env")
	envs, err := godotenv.Read(environmentPath)

	if err != nil {
		log.Fatalf("Error reading .env file")
	}
	return envs[key]
}
