package main

import (
	"database/sql"
	"github.com/bbalet/stopwords"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/reiver/go-porterstemmer"
	"gorm.io/gorm"
	html2 "html"
	"log"
	"strings"
)

type Index map[string][]uint

// normalize -> tokenize -> remove stop words -> stemming
func processText(text string, stem bool) []string {
	clean := stopwords.CleanString(text, "en", true)
	tokens := strings.Fields(clean)
	if stem {
		for i, _ := range tokens {
			tokens[i] = porterstemmer.StemString(tokens[i])
		}
	}

	return tokens
}

func safeRemoveHtml(text string) string {
	p := bluemonday.UGCPolicy()
	html := p.Sanitize(text)
	stripped := strip.StripTags(html)
	return html2.UnescapeString(stripped)
}

func (index Index) add(doc *Document) {
	titleTokens := processText(doc.Story.Title, true)
	for _, token := range titleTokens {
		index[token] = append(index[token], doc.Id)
	}

	for _, comment := range doc.Comments {
		commentTokens := processText(comment.Text, true)
		for _, token := range commentTokens {
			index[token] = append(index[token], doc.Id)
		}
	}
}

func (doc *Document) getTokens() (tokens map[string]int) {
	tokens = make(map[string]int)
	titleTokens := processText(doc.Story.Title, true)
	for _, token := range titleTokens {
		tokens[token] = tokens[token] + 1
	}

	for _, comment := range doc.Comments {
		commentTokens := processText(comment.Text, true)
		for _, token := range commentTokens {
			tokens[token] = tokens[token] + 1
		}
	}
	return tokens
}

func (comment *CommentModel) getCommentsTokens() (tokens map[string]int) {
	tokens = make(map[string]int)
	commentTokens := processText(comment.Text, true)
	for _, token := range commentTokens {
		tokens[token] = tokens[token] + 1
	}

	return tokens
}

func fetchAndIndexDocument(db *gorm.DB, id uint) {
	var documentModel DocumentModel
	err := db.First(&documentModel, id).Error
	if err == nil {
		log.Printf("Debug: Items with id %d already present in the index\n", id)
		return
	}

	doc, err := fetchStory(id)
	if err != nil {
		log.Printf("Error: Error %v while fetching item with id.\nError: %v\n", id, err)
		return
	}
	if doc.Story.Dead || doc.Story.Deleted {
		log.Printf("Debug: ignoring story since story: %d is either dead or deleted.\n", doc.Story.ID)
		return
	}

	if err := addDocumentToDbIndex(db, doc); err != nil {
		log.Printf("Error: Failed to add doc to db index titled: %s.\nError %v\n", doc.Story.Title, err)
		return
	}
}

func resumeHnIndexing(db *gorm.DB, backfill bool, maxDocumentCount int64) {

	var count int64
	if backfill {

		var maxID sql.NullInt64
		db.Model(&ResolvedItemModel{}).Select("MAX(id)").Scan(&maxID)

		if !maxID.Valid {
			id, err := fetchLatest()
			if err != nil {
				log.Printf("Error: Failed to find latest index\n")
				return
			}
			maxID.Int64 = int64(id)
		}

		for i := uint(maxID.Int64); i >= 1; i-- {

			db.Model(&ResolvedItemModel{}).Count(&count)
			if count >= maxDocumentCount {
				log.Printf("Info: Finished indexing %d items\n", maxDocumentCount)
				return
			}

			fetchAndIndexDocument(db, i)
		}

	} else {

		var minID sql.NullInt64
		db.Model(&ResolvedItemModel{}).Select("MIN(id)").Scan(&minID)
		if !minID.Valid {
			minID.Int64 = 1
		}

		for i := uint(minID.Int64); i >= 1; i-- {

			db.Model(&ResolvedItemModel{}).Count(&count)
			if count >= maxDocumentCount {
				log.Printf("Info: Finished indexing %d items\n", maxDocumentCount)
				return
			}

			fetchAndIndexDocument(db, i)
		}
	}
}
