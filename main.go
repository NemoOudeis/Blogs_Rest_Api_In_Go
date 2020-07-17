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
	return
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

func (blogs *Blogs) createBlogPost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodPost {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
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
		ExitWithError(response, statusCode, statusMessage)
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
		ExitWithError(response, statusCode, statusMessage)
	}

	docSnapshot, err := blogs.db.Collection("blogs").Doc(result.ID).Get(context.Background())
	if err != nil {
		statusCode := http.StatusInternalServerError
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
	}
	docSnapshotDatum := docSnapshot.Data()

	newBlogPost := BlogPost{
		ID:         result.ID,
		Title:      docSnapshotDatum["title"].(string),
		Content:    docSnapshotDatum["content"].(string),
		CreatedAt:  docSnapshotDatum["created_at"].(string),
		ModifiedAt: "",
	}

	statusCode := http.StatusCreated
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), newBlogPost)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
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
	router.HandleFunc("/blogs", blogs.getAllBlogPosts).Methods("GET")
	router.HandleFunc("/blogs/create", blogs.createBlogPost).Methods("POST")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))

}
