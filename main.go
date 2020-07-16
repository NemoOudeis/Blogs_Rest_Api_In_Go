package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

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

// Error is a structure which holds message for error, specifically in string json format
type Error struct {
	Message string `json:"message"`
}

func exitWithError(response http.ResponseWriter, statusCode int, statusMessage Error) {
	response.WriteHeader(statusCode)
	json.NewEncoder(response).Encode(statusMessage)
	return
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

func (blogs *Blogs) getAllBlogPosts(response http.ResponseWriter, request *http.Request) {
	greeting := "GetAllBlogPosts..."
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}

func (blogs *Blogs) createBlogPost(response http.ResponseWriter, request *http.Request) {
	greeting := "createBlogPost..."
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodPost {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		exitWithError(response, statusCode, statusMessage)
	}

	request.ParseForm()
	urlEncodedFormInputMap := request.Form
	title, isTitleFound := urlEncodedFormInputMap["title"]
	content, isContentFound := urlEncodedFormInputMap["content"]

	if isTitleFound == false || isContentFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		exitWithError(response, statusCode, statusMessage)
	}

	result, _, err := blogs.db.Collection("blogs").Add(context.Background(), map[string]interface{}{
		"title":      title,
		"content":    content,
		"created_at": time.Now(),
	})

	if err != nil {
		statusCode := http.StatusInternalServerError
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		exitWithError(response, statusCode, statusMessage)
	}

	// type createSuccessMessage struct {
	// 	id string
	// 	title string
	// 	content string
	// 	createdAt string
	// }
	//log.Println(result.ID, err)
	//json.NewDecoder(response.Body).Decode()

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

	blogs := initBlogs(client)

	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld).Methods("GET")
	router.HandleFunc("/blogs", blogs.getAllBlogPosts).Methods("GET")
	router.HandleFunc("/blogs/create", blogs.createBlogPost).Methods("POST")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))

}
