package main

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/api/iterator"
)

// HelloWorld greets the world for testing purpose
func HelloWorld(response http.ResponseWriter, request *http.Request) {
	greeting := "Hello World!"
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(greeting)
}

func (blogs *Blogs) getAllBlogPosts(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodGet {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
	}

	docSnapshotIter := blogs.db.Collection("blogs").Documents(context.Background())
	var allBlogPosts []BlogPost
	for {
		doc, err := docSnapshotIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			statusCode := http.StatusInternalServerError
			statusMessage := Error{
				Message: http.StatusText(statusCode),
			}
			ExitWithError(response, statusCode, statusMessage)
		}

		docSnapshotDatum := doc.Data()

		optionalModifiedField := docSnapshotDatum["modified_at"]
		var ModifiedField string
		if optionalModifiedField != nil {
			ModifiedField = docSnapshotDatum["modified_at"].(string)
		} else {
			ModifiedField = ""
		}

		blogPost := BlogPost{
			ID:         doc.Ref.ID,
			Title:      docSnapshotDatum["title"].(string),
			Content:    docSnapshotDatum["content"].(string),
			CreatedAt:  docSnapshotDatum["created_at"].(string),
			ModifiedAt: ModifiedField,
		}
		allBlogPosts = append(allBlogPosts, blogPost)
	}

	statusCode := http.StatusOK
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), allBlogPosts)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}
