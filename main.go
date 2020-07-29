package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// LoadEnvFileAndReturnEnvVarValueByKey returns value of given variable inside .env
func LoadEnvFileAndReturnEnvVarValueByKey(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

// Error is a structure which holds message for error in string json format
type Error struct {
	Message       string `json:"message"`
	CustomMessage string `json:"custom_message,omitempty"`
}

// Env holds global environment variable to configure environment within this app
type Env struct {
	Port              int
	FirebaseProjectID string
	JwtHashKey        string
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

var env = Env{
	Port:              8081,
	FirebaseProjectID: LoadEnvFileAndReturnEnvVarValueByKey("FIREBASE_PROJECT_ID"),
	JwtHashKey:        LoadEnvFileAndReturnEnvVarValueByKey("JWT_HASH_KEY")}

func main() {

	ctx := context.Background()
	sa := option.WithCredentialsFile("yurie-s-go-api-firebase-adminsdk-qzfyx-d2587d9fd3.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error getting specificed app from firebase: %v\n", err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error getting Firestore client: %v\n", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	blogs := initBlogs(firestoreClient)
	users := initUsers(firestoreClient, authClient)

	router := mux.NewRouter()

	router.HandleFunc("/", HelloWorld).Methods("GET")
	router.HandleFunc("/signup", users.Signup)
	router.HandleFunc("/login", users.Login)
	router.HandleFunc("/blogs", users.verifyToken(blogs.ListAllArticlesHandler))
	router.HandleFunc("/blogs/create", users.verifyToken(blogs.PublishArticleHandler))
	router.HandleFunc("/blogs/{id}", users.verifyToken(blogs.ListArticleHandler))
	router.HandleFunc("/blogs/delete/{id}", users.verifyToken(blogs.DeleteArticleHandler))
	router.HandleFunc("/blogs/update/{id}", users.verifyToken(blogs.UpdateArticleHandler))

	defer firestoreClient.Close()

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", env.Port), router))
}
