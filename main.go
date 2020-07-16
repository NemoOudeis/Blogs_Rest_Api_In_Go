package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func loadEnvFileAndReturnEnvVarValueByKey(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

// Blogs is a structure which holds database and handler for database operation over HTTP calls
type Blogs struct {
	db *firestore.Client
}

func createFirestoreClient(firebaseContext context.Context) *firestore.Client {
	projectID := loadEnvFileAndReturnEnvVarValueByKey("FIREBASE_PROJECT_ID")
	firestoreClient, err := firestore.NewClient(firebaseContext, projectID)
	if err != nil {
		log.Fatalf("Failed to create client firestore: %v", err)
	}
	return firestoreClient
}

func initBlogs(db *firestore.Client) *Blogs {
	return &Blogs{db: db}
}

func main() {

	firebaseContext := context.Background()
	db := createFirestoreClient(firebaseContext)
	defer db.Close()

	blogs := initBlogs(db)

	router := mux.NewRouter()

	router.HandleFunc("/", HelloWorld).Methods("GET")
	router.HandleFunc("/blogs", GetAllBlogPosts).Methods("GET")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))

}
