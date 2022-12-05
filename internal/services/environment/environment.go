package environment

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadGoDotEnv to load the .env file into application environment
func LoadGoDotEnv() {
	// load .env file
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	environmentPath := filepath.Join(dir, "config/.env")
	log.Println(environmentPath)
	if err := godotenv.Load(environmentPath); err != nil {
		log.Println("No .env file found")
	}
}
