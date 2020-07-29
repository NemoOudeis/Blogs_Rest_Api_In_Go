package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func (blogs *Blogs) getAllArticles() ([]*Article, error) {
	var articles []*Article

	docSnapshotIter := blogs.db.Collection("blogs").Documents(context.Background())
	for {
		doc, err := docSnapshotIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		docSnapshotDatum := doc.Data()

		optionalModifiedField := docSnapshotDatum["modified_at"]
		var ModifiedField string
		if optionalModifiedField != nil {
			ModifiedField = docSnapshotDatum["modified_at"].(string)
		} else {
			ModifiedField = ""
		}

		article := Article{
			ID:         doc.Ref.ID,
			Title:      docSnapshotDatum["title"].(string),
			Content:    docSnapshotDatum["content"].(string),
			CreatedAt:  docSnapshotDatum["created_at"].(string),
			ModifiedAt: ModifiedField,
		}
		articles = append(articles, &article)
	}
	return articles, nil
}

// GetArticleByID gets existing article from the DB by given ID
func (blogs *Blogs) GetArticleByID(ID string) (*Article, error) {
	docSnapshot, err := blogs.db.Collection("blogs").Doc(ID).Get(context.Background())
	if err != nil {
		return nil, err
	}
	docSnapshotDatum := docSnapshot.Data()

	optionalField := docSnapshotDatum["modified_at"]
	var modifiedTimeSlot string
	if optionalField != nil {
		modifiedTimeSlot = docSnapshotDatum["modified_at"].(string)
	} else {
		modifiedTimeSlot = ""
	}

	article := Article{
		ID:         docSnapshot.Ref.ID,
		Title:      docSnapshotDatum["title"].(string),
		Content:    docSnapshotDatum["content"].(string),
		CreatedAt:  docSnapshotDatum["created_at"].(string),
		ModifiedAt: modifiedTimeSlot,
	}
	return &article, nil
}

// AddArticle adds a new article to the DB with given title and content
func (blogs *Blogs) AddArticle(title, content string) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	result, _, err := blogs.db.Collection("blogs").Add(context.Background(), map[string]interface{}{
		"title":      title,
		"content":    content,
		"created_at": time.Now().String(),
	})
	return result, nil, err
}
