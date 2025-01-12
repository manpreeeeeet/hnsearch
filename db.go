package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBCreds struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     string
}

func getConnection(creds DBCreds) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", creds.Host, creds.User, creds.Password, creds.DBName, creds.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

type TokenModel struct {
	gorm.Model
	Token     string
	Documents []DocumentModel `gorm:"many2many:document_token_frequency_models;joinForeignKey:token_id;joinReferences:document_id"`
}

type DocumentModel struct {
	gorm.Model
	ID       uint
	Title    string
	Tokens   []TokenModel `gorm:"many2many:document_token_frequency_models;joinForeignKey:document_id;joinReferences:token_id"`
	Comments []CommentModel
}

type CommentModel struct {
	gorm.Model
	Text            string
	DocumentModelID uint
}

type DocumentTokenFrequencyModel struct {
	TokenID    uint
	DocumentID uint
	Frequency  int
}

func loadTokensToIndex(db *gorm.DB) (Index, error) {
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

func (doc *Document) toDocumentModel() *DocumentModel {
	comments := make([]CommentModel, 0)
	for _, comment := range doc.Comments {
		comments = append(comments, CommentModel{DocumentModelID: doc.Id, Text: safeRemoveHtml(comment.Text)})
	}
	return &DocumentModel{
		ID:       doc.Id,
		Title:    safeRemoveHtml(doc.Story.Title),
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

func addDocumentToDbIndex(db *gorm.DB, doc *Document) error {
	documentModel := *doc.toDocumentModel()
	db.Create(&documentModel)

	tokens := doc.getTokens()
	tokensModel, err := addTokensWithFrequency(db, doc.Id, tokens)
	if err != nil {
		return err
	}

	documentModel.Tokens = tokensModel
	if err := db.Save(documentModel).Error; err != nil {
		fmt.Printf("failed to associate terms with document: %s", err)
		return err
	}

	return nil
}

func addTokensWithFrequency(db *gorm.DB, docID uint, tokens map[string]int) (tokensModel []TokenModel, err error) {
	for token, freq := range tokens {
		tokenModel := TokenModel{Token: token}
		if err := db.Where("token = ?", token).FirstOrCreate(&tokenModel).Error; err != nil {
			return nil, err
		}

		if err := db.Create(&DocumentTokenFrequencyModel{
			TokenID:    tokenModel.ID,
			DocumentID: docID,
			Frequency:  freq,
		}).Error; err != nil {
			return nil, err
		}
		tokensModel = append(tokensModel, tokenModel)
	}
	return tokensModel, nil
}

func getTokenModel(db *gorm.DB, token string) *TokenModel {
	tokenModel := TokenModel{Token: token}
	if err := db.Where("token = ?", token).FirstOrCreate(&tokenModel).Error; err != nil {
		//log.Fatalf("Error getting token: %s\n", token)
		return nil
	}
	return &tokenModel
}

func getTokenFrequency(db *gorm.DB, docID uint, tokenID uint) int64 {
	var frequencyRecord DocumentTokenFrequencyModel
	err := db.Where("token_id = ? AND document_id = ?", tokenID, docID).First(&frequencyRecord).Error
	if err != nil {
		return 0
	}
	return int64(frequencyRecord.Frequency)
}

func getInverseDocumentFrequency(db *gorm.DB, tokenID uint) int64 {
	var docFreq int64
	err := db.Model(&DocumentTokenFrequencyModel{}).Where("token_id = ?", tokenID).Count(&docFreq).Error
	if err != nil {
		//log.Fatalf("Error getting document frequency for token %d: %v", tokenID, err)
		return 0
	}
	return docFreq
}
