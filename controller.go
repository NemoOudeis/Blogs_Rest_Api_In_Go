package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
)

// HelloWorld greets the world for testing purpose
func HelloWorld(response http.ResponseWriter, request *http.Request) {
	greeting := "Hello World!"
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}

// Blogs is a structure which holds database and handler for database operation over HTTP calls
type Blogs struct {
	db *firestore.Client
}

// Article is a standard format of single blog post data (document snapshot)
type Article struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at,omitempty"`
}

func initBlogs(db *firestore.Client) *Blogs {
	return &Blogs{db: db}
}

// ListAllArticles lists all articles available inside the DB
func (blogs *Blogs) ListAllArticles(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodGet {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	allArticles, err := blogs.getAllArticles()
	if err != nil {
		statusCode := http.StatusInternalServerError
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}
	statusCode := http.StatusOK
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), allArticles)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}

// PublishArticle publishes an article with given title and content
func (blogs *Blogs) PublishArticle(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

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
	title, isTitleFound := urlEncodedFormInputMap["title"]
	content, isContentFound := urlEncodedFormInputMap["content"]

	if isTitleFound == false || isContentFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
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
		return
	}

	docSnapshot, err := blogs.db.Collection("blogs").Doc(result.ID).Get(context.Background())
	if err != nil {
		statusCode := http.StatusInternalServerError
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}
	docSnapshotDatum := docSnapshot.Data()

	newArticle := Article{
		ID:         result.ID,
		Title:      docSnapshotDatum["title"].(string),
		Content:    docSnapshotDatum["content"].(string),
		CreatedAt:  docSnapshotDatum["created_at"].(string),
		ModifiedAt: "",
	}

	statusCode := http.StatusCreated
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), newArticle)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}

// ListArticleByID lists an article by ID
func (blogs *Blogs) ListArticleByID(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodGet {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	param := mux.Vars(request)
	ID := param["id"]

	if len(ID) == 0 {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	docSnapshot, err := blogs.db.Collection("blogs").Doc(ID).Get(context.Background())
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	docSnapshotDatum := docSnapshot.Data()

	optionalModifiedField := docSnapshotDatum["modified_at"]
	var ModifiedField string
	if optionalModifiedField != nil {
		ModifiedField = docSnapshotDatum["modified_at"].(string)
	} else {
		ModifiedField = ""
	}

	article := Article{
		ID:         docSnapshot.Ref.ID,
		Title:      docSnapshotDatum["title"].(string),
		Content:    docSnapshotDatum["content"].(string),
		CreatedAt:  docSnapshotDatum["created_at"].(string),
		ModifiedAt: ModifiedField,
	}

	statusCode := http.StatusOK
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), article)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}

// DeleteArticleByID deletes an article by ID
func (blogs *Blogs) DeleteArticleByID(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodDelete {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	param := mux.Vars(request)
	ID := param["id"]
	if len(ID) == 0 {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	_, err := blogs.db.Collection("blogs").Doc(ID).Get(context.Background())
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	_, err = blogs.db.Collection("blogs").Doc(ID).Delete(context.Background())
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	customMessage := fmt.Sprintf("The Blog post with ID %s was successfully deleted.", ID)

	statusCode := http.StatusOK
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), customMessage)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}

// UpdateArticleByID updates an article by ID
func (blogs *Blogs) UpdateArticleByID(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodPut {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
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
		return
	}

	param := mux.Vars(request)
	ID := param["id"]
	if len(ID) == 0 {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	ref := blogs.db.Collection("blogs").Doc(ID)
	err := blogs.db.RunTransaction(context.Background(), func(ctx context.Context, tx *firestore.Transaction) error {
		_, err := tx.Get(ref)
		if err != nil {
			statusCode := http.StatusServiceUnavailable
			statusMessage := Error{
				// err.Error() is a custom error message from client firestore API
				Message: err.Error(),
			}
			ExitWithError(response, statusCode, statusMessage)
			return nil
		}

		return tx.Set(ref, map[string]interface{}{
			"title":       strings.Join(title, ""),
			"content":     strings.Join(content, ""),
			"modified_at": time.Now().String(),
		}, firestore.MergeAll)
	})
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
	}

	customMessage := fmt.Sprintf("The Blog post with ID %s was successfully updated.", ID)

	statusCode := http.StatusOK
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), customMessage)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}
