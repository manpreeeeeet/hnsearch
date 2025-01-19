package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"math"
	"time"
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

type ResolvedItemModel struct {
	gorm.Model
	ID uint
}

type TokenModel struct {
	Token     string          `gorm:"primaryKey"`
	Documents []DocumentModel `gorm:"many2many:document_token_frequency_models;joinForeignKey:token;joinReferences:document_id"`
}

type DocumentModel struct {
	gorm.Model
	ID       uint
	URL      string
	Score    int
	Title    string
	Tokens   []TokenModel `gorm:"many2many:document_token_frequency_models;joinForeignKey:document_id;joinReferences:token"`
	Comments []CommentModel
}

type CommentModel struct {
	gorm.Model
	ID              uint
	Text            string
	DocumentModelID uint
}

type CommentTokenFrequencyModel struct {
	Token      string `gorm:"primaryKey"`
	CommentID  uint   `gorm:"primaryKey"`
	DocumentID uint   `gorm:"primaryKey"`
	Frequency  int
}

type DocumentTokenFrequencyModel struct {
	Token      string `gorm:"primaryKey"`
	DocumentID uint   `gorm:"primaryKey"`
	Frequency  int
}

type DocumentTokenCount struct {
	DocumentID  uint `gorm:"primaryKey"`
	TotalTokens int
}

func (DocumentTokenCount) TableName() string {
	return "document_token_counts_view"
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
		comments = append(comments, CommentModel{ID: comment.ID, DocumentModelID: doc.Id, Text: safeRemoveHtml(comment.Text)})
	}
	return &DocumentModel{
		ID:       doc.Id,
		URL:      doc.Story.URL,
		Title:    safeRemoveHtml(doc.Story.Title),
		Comments: comments,
		Score:    doc.Story.Score,
	}
}

func (doc *DocumentModel) toDocument() Document {
	comments := make([]Comment, 0)
	for _, comment := range doc.Comments {
		comments = append(comments, Comment{Text: comment.Text})
	}
	return Document{
		Id:       doc.ID,
		Story:    Story{Title: doc.Title, URL: doc.URL, Score: doc.Score},
		Comments: comments,
	}
}

func addDocumentToDbIndex(db *gorm.DB, doc *Document) error {

	return db.Transaction(func(tx *gorm.DB) error {
		documentModel := *doc.toDocumentModel()
		tx.Create(&documentModel)

		tokens := doc.getTokens()
		if _, err := addTokensWithFrequency(tx, doc.Id, tokens); err != nil {
			return err
		}
		if _, err := addCommentTokensWithFrequency(tx, documentModel); err != nil {
			return err
		}

		resolvedItems := []*ResolvedItemModel{
			{ID: doc.Id},
		}
		for _, comment := range documentModel.Comments {
			resolvedItems = append(resolvedItems, &ResolvedItemModel{ID: comment.ID})
		}
		tx.Create(resolvedItems)

		return nil
	})

}

func addTokensWithFrequency(db *gorm.DB, docID uint, tokens map[string]int) (tokensModel []TokenModel, err error) {
	for token, freq := range tokens {
		tokenModel := TokenModel{Token: token}
		if err := db.Where("token = ?", token).FirstOrCreate(&tokenModel).Error; err != nil {
			return nil, err
		}

		if err := db.Create(&DocumentTokenFrequencyModel{
			Token:      token,
			DocumentID: docID,
			Frequency:  freq,
		}).Error; err != nil {
			return nil, err
		}
		tokensModel = append(tokensModel, tokenModel)
	}
	return tokensModel, nil
}

func addCommentTokensWithFrequency(db *gorm.DB, doc DocumentModel) (tokensModel []TokenModel, err error) {

	for _, commentModel := range doc.Comments {
		tokens := commentModel.getCommentsTokens()

		for token, freq := range tokens {
			tokenModel := TokenModel{Token: token}
			if err := db.Where("token = ?", token).FirstOrCreate(&tokenModel).Error; err != nil {
				return nil, err
			}

			if err := db.Create(&CommentTokenFrequencyModel{
				Token:      token,
				CommentID:  commentModel.ID,
				DocumentID: doc.ID,
				Frequency:  freq,
			}).Error; err != nil {
				return nil, err
			}
			tokensModel = append(tokensModel, tokenModel)
		}
	}
	return tokensModel, nil
}

func getInverseDocumentFrequencies(db *gorm.DB, tokens []string) (map[string]float64, error) {
	start := time.Now()
	var totalDocs int64
	if err := db.Model(&DocumentModel{}).Count(&totalDocs).Error; err != nil {
		return nil, err
	}

	type TokenCount struct {
		Token   string
		TokenID uint
		DocFreq int64
	}
	var tokenCounts []TokenCount

	dtfStart := time.Now()
	err := db.Model(&DocumentTokenFrequencyModel{}).
		Select("token_id, token, COUNT(DISTINCT document_id) as doc_freq").
		Joins("JOIN token_models tm ON tm.id = document_token_frequency_models.token_id").
		Where("tm.token IN ?", tokens).
		Group("token_id, tm.token").
		Scan(&tokenCounts).Error
	log.Printf("document token freq took %s", time.Since(dtfStart))

	if err != nil {
		return nil, err
	}

	idfs := make(map[string]float64)
	for _, tc := range tokenCounts {
		if tc.DocFreq > 0 {
			idfs[tc.Token] = math.Log(float64(totalDocs) / (1 + float64(tc.DocFreq)))
		}
	}
	elapsed := time.Since(start)
	log.Printf("get inverse freq took %s", elapsed)
	return idfs, nil
}

func getNormalizedTokenFrequencies(db *gorm.DB, docIDs []uint, tokens []string) (map[uint]map[string]float64, error) {

	type TokenFreq struct {
		DocumentID uint
		Token      string
		Frequency  int
	}
	var tokenFreqs []TokenFreq

	dtfModelStart := time.Now()
	err := db.Debug().Table("document_token_frequency_models").
		Select("document_token_frequency_models.document_id, token_models.token, document_token_frequency_models.frequency").
		Joins("INNER JOIN token_models ON token_models.id = document_token_frequency_models.token_id").
		Where("document_id IN ? AND token_models.token IN ?", docIDs, tokens).
		Scan(&tokenFreqs).Error
	log.Printf("document token frequency took %s", time.Since(dtfModelStart))

	if err != nil {
		return nil, err
	}

	// Get total frequencies per document in one query
	type DocTotal struct {
		DocumentID  uint
		TotalTokens int
	}
	var docTotals []DocTotal

	documentTotalStart := time.Now()
	//err = db.Debug().Model(&DocTotal{}).Where("document_id")
	err = db.Debug().Table("document_token_frequency_models").
		Select("document_id, SUM(frequency) as total_tokens").
		Where("document_id IN ?", docIDs).
		Group("document_id").
		Scan(&docTotals).Error
	log.Printf("documentTotalStart took %s", time.Since(documentTotalStart))

	if err != nil {
		return nil, err
	}

	totalTokensMap := make(map[uint]int)
	for _, dt := range docTotals {
		totalTokensMap[dt.DocumentID] = dt.TotalTokens
	}

	result := make(map[uint]map[string]float64)
	for _, tf := range tokenFreqs {
		if totalTokensMap[tf.DocumentID] > 0 {
			if result[tf.DocumentID] == nil {
				result[tf.DocumentID] = make(map[string]float64)
			}
			result[tf.DocumentID][tf.Token] = float64(tf.Frequency) / (1 + float64(totalTokensMap[tf.DocumentID]))
		}
	}

	return result, nil
}

func MigrateToTokenPrimaryKey(db *gorm.DB) error {
	// Step 1: Create new tables with token as primary key
	err := db.AutoMigrate(
		&TokenModel{},
		&DocumentModel{},
		&CommentModel{},
		&DocumentTokenFrequencyModel{},
		&CommentTokenFrequencyModel{},
	)
	if err != nil {
		return fmt.Errorf("failed to create new tables: %v", err)
	}

	// Step 2: Create temporary tables for migration
	err = db.Exec(`
        CREATE TABLE temp_document_token_frequency AS
        SELECT 
            tm.token,
            dtf.document_id,
            dtf.frequency
        FROM document_token_frequency_models dtf
        JOIN token_models tm ON tm.id = dtf.token_id;

        CREATE TABLE temp_comment_token_frequency AS
        SELECT 
            tm.token,
            ctf.comment_id,
            ctf.document_id,
            ctf.frequency
        FROM comment_token_frequency_models ctf
        JOIN token_models tm ON tm.id = ctf.token_id;
    `).Error
	if err != nil {
		return fmt.Errorf("failed to create temp tables: %v", err)
	}

	// Step 3: Drop old tables and constraints
	err = db.Exec(`
        DROP TABLE IF EXISTS document_token_frequency_models;
        DROP TABLE IF EXISTS comment_token_frequency_models;
        DROP TABLE IF EXISTS token_models CASCADE;
    `).Error
	if err != nil {
		return fmt.Errorf("failed to drop old tables: %v", err)
	}

	err = db.AutoMigrate(
		&TokenModel{},
		&DocumentModel{},
		&CommentModel{},
		&DocumentTokenFrequencyModel{},
		&CommentTokenFrequencyModel{},
	)
	if err != nil {
		return fmt.Errorf("failed to create new tables: %v", err)
	}

	// Step 4: Migrate data from temporary tables to new tables
	err = db.Exec(`
        INSERT INTO token_models (token, created_at, updated_at)
        SELECT DISTINCT token, NOW(), NOW()
        FROM temp_document_token_frequency;

        INSERT INTO document_token_frequency_models (token, document_id, frequency)
        SELECT token, document_id, frequency
        FROM temp_document_token_frequency;

        INSERT INTO comment_token_frequency_models (token, comment_id, document_id, frequency)
        SELECT token, comment_id, document_id, frequency
        FROM temp_comment_token_frequency;

        DROP TABLE temp_document_token_frequency;
        DROP TABLE temp_comment_token_frequency;
    `).Error
	if err != nil {
		return fmt.Errorf("failed to migrate data: %v", err)
	}

	// Step 5: Create necessary indexes
	err = db.Exec(`
        CREATE INDEX idx_document_token_freq ON document_token_frequency_models(token, document_id);
        CREATE INDEX idx_comment_token_freq ON comment_token_frequency_models(token, comment_id, document_id);
    `).Error
	if err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	return nil
}
