package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	greeting := "Hello World!"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(greeting)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld).Methods("GET")
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8081", router))
}
