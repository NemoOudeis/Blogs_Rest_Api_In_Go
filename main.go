package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/option"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	firebase "firebase.google.com/go"
)

func helloWorld(response http.ResponseWriter, request *http.Request) {
	greeting := "Hello World!"
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}

func loadEnvFileAndReturnEnvVarValueByKey(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func initFirebaseClientSDK() {
	firebaseContext := context.Background()
	firebaseServiceAccount := option.WithCredentialsFile("yurie-s-go-api-firebase-adminsdk-qzfyx-d2587d9fd3.json")
	firebaseApp, err := firebase.NewApp(firebaseContext, nil, firebaseServiceAccount)
	if err != nil {
		log.Fatalf("Error initializing firebase app: %v", err)
	}

	firestore, err := firebaseApp.Firestore(firebaseContext)
	if err != nil {
		log.Fatalf("Error initializing firestore: %v", err)
	}

	defer firestore.Close()
}

func main() {

	initFirebaseClientSDK()
	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld).Methods("GET")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))

}
