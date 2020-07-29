package main

import (
	"context"

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
