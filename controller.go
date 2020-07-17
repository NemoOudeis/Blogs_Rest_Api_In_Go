package main

import (
	"encoding/json"
	"net/http"
)

// HelloWorld greets the world for testing purpose
func HelloWorld(response http.ResponseWriter, request *http.Request) {
	greeting := "Hello World!"
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}
