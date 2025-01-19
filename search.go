package main

import (
	"gorm.io/gorm"
	"log"
	"sort"
	"time"
)

func searchQuery(db *gorm.DB, query string) []DocumentModel {
	tokens := processText(query, true)
	documents := searchDocuments(db, tokens)
	documents = rankDocuments(db, tokens, documents)
	return documents
}

func searchDocuments(db *gorm.DB, tokens []string) []DocumentModel {
	start := time.Now()

	var documents []DocumentModel
	err := db.Debug().Model(&DocumentModel{}).
		Joins("INNER JOIN document_token_frequency_models dtf ON dtf.document_id = document_models.id AND dtf.token IN ?", tokens).
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Debug().
				Distinct().
				Joins("INNER JOIN comment_token_frequency_models ctf ON ctf.comment_id = comment_models.id").
				Where("ctf.token IN ? AND ctf.document_id = comment_models.document_model_id", tokens)
		}).
		Find(&documents).Error //err = db.

	if err != nil {
		// handle error
	}
	elapsed := time.Since(start)
	log.Printf("search docs took %s", elapsed)
	return documents

}

func rankDocuments(db *gorm.DB, tokens []string, docs []DocumentModel) []DocumentModel {
	start := time.Now()
	// find frequency of tokens in each document
	// ivf (how rare this thing is) -> count(id) where token = ?
	//tokenFreqWeight := 0.1
	//inverseFreqWeight := 0.4

	inverseTokenFreq, _ := getInverseDocumentFrequencies(db, tokens)
	docScore := map[uint]float64{}

	docIDs := make([]uint, 0)
	for _, doc := range docs {
		docIDs = append(docIDs, doc.ID)
	}
	normalizedFreqs, err := getNormalizedTokenFrequencies(db, docIDs, tokens)
	if err != nil {
		// handle error
	}

	for docID, tokenFreqs := range normalizedFreqs {
		for token, freq := range tokenFreqs {
			docScore[docID] += freq * inverseTokenFreq[token]
		}
	}

	sort.Slice(docs, func(i, j int) bool {
		return docScore[docs[i].ID] > docScore[docs[j].ID]
	})

	elapsed := time.Since(start)
	log.Printf("rank docs took %s", elapsed)
	return docs
}
