package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
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

// Users is a structure which holds database for CRUD operation in the client and app initialized for admin in the backend
type Users struct {
	db         *firestore.Client
	authClient *auth.Client
}

// User holds basic user info of a current user
type User struct {
	UID      string `json:"uid"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func initUsers(db *firestore.Client, authClient *auth.Client) *Users {
	return &Users{db: db, authClient: authClient}
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
	Message       string `json:"message"`
	CustomMessage string `json:"custom_message,omitempty"`
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

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	defer client.Close()

	blogs := initBlogs(client)
	users := initUsers(client, authClient)

	router := mux.NewRouter()

	router.HandleFunc("/", HelloWorld).Methods("GET")

	router.HandleFunc("/signup", users.signup)

	router.HandleFunc("/blogs", blogs.getAllBlogPosts)
	router.HandleFunc("/blogs/create", blogs.createBlogPost)
	router.HandleFunc("/blogs/{id}", blogs.getBlogPostByID)
	router.HandleFunc("/blogs/delete/{id}", blogs.deleteBlogPostByID)
	router.HandleFunc("/blogs/update/{id}", blogs.updateBlogPostByID)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))
}

func (users *Users) signup(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	request.ParseForm()
	urlEncodedFormInputMap := request.Form
	email, isEmailFound := urlEncodedFormInputMap["email"]
	password, isPasswordFound := urlEncodedFormInputMap["password"]

	if isEmailFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Email is required.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	if isPasswordFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Password is required.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(strings.Join(email, "")).
		Password(strings.Join(password, "")).
		Disabled(false)

	newUser, err := users.authClient.CreateUser(context.Background(), params)
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message:       err.Error(),
			CustomMessage: "error creating a user",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	log.Printf("Successfully created user: %v\n", newUser.UserInfo)

}
