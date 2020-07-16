package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
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

func helloWorld(response http.ResponseWriter, request *http.Request) {
	greeting := "Hello World!"
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}

func getAllBlogPosts(response http.ResponseWriter, request *http.Request) {
	greeting := "GetAllBlogPosts..."
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}

func main() {

	ctx := context.Background()
	sa := option.WithCredentialsFile("yurie-s-go-api-firebase-adminsdk-qzfyx-d2587d9fd3.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	result, err := client.Collection("blogs").Doc("test").Set(context.Background(), map[string]interface{}{
		"first": "Ada",
		"last":  "Lovelace",
		"born":  1815,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(result)

	defer client.Close()

	//blogs := initBlogs(db)

	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld).Methods("GET")
	router.HandleFunc("/blogs", getAllBlogPosts).Methods("GET")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))

}
