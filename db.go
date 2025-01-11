package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getConnection() (*gorm.DB, error) {
	dsn := "host=localhost user=testuser password=testpassword dbname=testdb port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

type TokenModel struct {
	gorm.Model
	Token     string
	Documents []DocumentModel `gorm:"many2many:term_documents"`
}

func loadTokensToMap(db *gorm.DB) (Index, error) {
	tokenToDocs := make(map[string][]uint)

	var tokenModels []TokenModel
	if err := db.Preload("Documents").Find(&tokenModels).Error; err != nil {
		return nil, fmt.Errorf("failed to load token models: %w", err)
	}

	for _, tokenModel := range tokenModels {
		var docIDs []uint
		for _, document := range tokenModel.Documents {
			docIDs = append(docIDs, document.ID)
		}
		tokenToDocs[tokenModel.Token] = docIDs
	}

	return tokenToDocs, nil
}

// these are only storing the raw data
type DocumentModel struct {
	gorm.Model
	ID       uint
	Title    string
	Tokens   []TokenModel `gorm:"many2many:term_documents"`
	Comments []CommentModel
}

type CommentModel struct {
	gorm.Model
	Text            string
	DocumentModelID uint
}

func (doc *Document) toDocumentModel() *DocumentModel {
	comments := make([]CommentModel, 0)
	for _, comment := range doc.Comments {
		comments = append(comments, CommentModel{DocumentModelID: doc.Id, Text: comment.Text})
	}
	return &DocumentModel{
		ID:       doc.Id,
		Title:    doc.Story.Title,
		Comments: comments,
	}
}

func (doc *DocumentModel) toDocument() Document {
	comments := make([]Comment, 0)
	for _, comment := range doc.Comments {
		comments = append(comments, Comment{Text: comment.Text})
	}
	return Document{
		Id:       doc.ID,
		Story:    Story{Title: doc.Title},
		Comments: comments,
	}
}
