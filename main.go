package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
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

// CreatedBlogPost holds fields which newly created data holds
type CreatedBlogPost struct {
	ID        string
	Title     string
	Content   string
	CreatedAt string
}

// Error is a structure which holds message for error in string json format
type Error struct {
	Message string `json:"message"`
}

// Success is a structure which holds message for successful db operation in string json format
type Success struct {
	Message     string          `json:"message"`
	NewBlogPost CreatedBlogPost `json:"newBlogPost"`
}

func exitWithError(response http.ResponseWriter, statusCode int, statusMessage Error) {
	response.WriteHeader(statusCode)
	json.NewEncoder(response).Encode(statusMessage)
	return
}

func returnSuccessfulResponse(response http.ResponseWriter, statusCode int, statusMessage Success) {
	response.WriteHeader(statusCode)
	json.NewEncoder(response).Encode(statusMessage)
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
		"title":      strings.Join(title, ""),
		"content":    strings.Join(content, ""),
		"created_at": time.Now().String(),
	})

	if err != nil {
		statusCode := http.StatusInternalServerError
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		exitWithError(response, statusCode, statusMessage)
	}

	docSnapshot, err := blogs.db.Collection("blogs").Doc(result.ID).Get(context.Background())
	if err != nil {
		statusCode := http.StatusInternalServerError
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		exitWithError(response, statusCode, statusMessage)
	}
	docSnapshotDatum := docSnapshot.Data()

	newBlogPost := CreatedBlogPost{
		ID:        result.ID,
		Title:     docSnapshotDatum["title"].(string),
		Content:   docSnapshotDatum["content"].(string),
		CreatedAt: docSnapshotDatum["created_at"].(string),
	}

	statusCode := http.StatusCreated
	statusMessage := Success{
		Message:     http.StatusText(statusCode),
		NewBlogPost: newBlogPost,
	}
	returnSuccessfulResponse(response, statusCode, statusMessage)
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

	router.HandleFunc("/", helloWorld).Methods("GET")
	router.HandleFunc("/blogs", blogs.getAllBlogPosts).Methods("GET")
	router.HandleFunc("/blogs/create", blogs.createBlogPost).Methods("POST")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))

}
