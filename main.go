package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func helloWorld(response http.ResponseWriter, request *http.Request) {
	greeting := "Hello World!"
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld).Methods("GET")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))
}
