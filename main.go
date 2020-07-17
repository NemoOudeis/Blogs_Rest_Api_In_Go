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

// BlogPost is a standard format of single blog post data (document snapshot)
type BlogPost struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at,omitempty"`
}

// Error is a structure which holds message for error in string json format
type Error struct {
	Message string `json:"message"`
}

// ExitWithError exits from a function when any type of err was caught during http communication
func ExitWithError(response http.ResponseWriter, statusCode int, statusMessage Error) {
	response.WriteHeader(statusCode)
	json.NewEncoder(response).Encode(statusMessage)
}

// ReturnSuccessfulResponse returns a success message to the client
func ReturnSuccessfulResponse(response http.ResponseWriter, statusCode int, statusMessage map[string]interface{}) {
	response.WriteHeader(statusCode)
	json.NewEncoder(response).Encode(statusMessage)
}

// SuccessJSONGenerator is a factory to generate message when job is successfully finished over http connection
func SuccessJSONGenerator(msgVal, dataVal interface{}) map[string]interface{} {
	return map[string]interface{}{
		"Message": msgVal,
		"Data":    dataVal,
	}
}

func initBlogs(db *firestore.Client) *Blogs {
	return &Blogs{db: db}
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

	defer client.Close()

	blogs := initBlogs(client)

	router := mux.NewRouter()

	router.HandleFunc("/", HelloWorld).Methods("GET")
	router.HandleFunc("/blogs", blogs.getAllBlogPosts)
	router.HandleFunc("/blogs/create", blogs.createBlogPost)
	router.HandleFunc("/blogs/{id}", blogs.getBlogPostByID)
	router.HandleFunc("/blogs/delete/{id}", blogs.deleteBlogPostByID)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))
}
