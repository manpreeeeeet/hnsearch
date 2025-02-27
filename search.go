package main

import (
	"gorm.io/gorm"
	"log"
	"sort"
	"time"
)

func searchQuery(db *gorm.DB, query string, docTotals []DocumentTokenCount) []DocumentModel {
	tokens := processText(query, true)
	docIds := rankDocuments(db, tokens, docTotals)
	documents := searchDocuments(db, docIds, tokens)
	return documents
}

func searchDocuments(db *gorm.DB, docIds []uint, tokens []string) []DocumentModel {
	start := time.Now()

	var documents []DocumentModel
	err := db.Debug().Model(&DocumentModel{}).
		Where("id IN ?", docIds).
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Debug().
				Where("comment_models.deleted_at IS NULL").
				Joins("INNER JOIN comment_token_frequency_models ctf ON ctf.comment_id = comment_models.id").
				Where("ctf.token IN ? AND ctf.document_id = comment_models.document_model_id", tokens).
				Distinct()
		}).
		Find(&documents).Error //err = db.

	if err != nil {
		// handle error
	}
	elapsed := time.Since(start)
	log.Printf("search docs took %s", elapsed)
	return documents

}

func rankDocuments(db *gorm.DB, tokens []string, docTotals []DocumentTokenCount) []uint {
	start := time.Now()
	// find frequency of tokens in each document
	// ivf (how rare this thing is) -> count(id) where token = ?
	//tokenFreqWeight := 0.1
	//inverseFreqWeight := 0.4
	inverseTokenFreq, _ := getInverseDocumentFrequencies(db, tokens)
	docScore := map[uint]float64{}

	normalizedFreqs, err := getNormalizedTokenFrequencies(db, tokens, docTotals)
	if err != nil {
		// handle error
	}

	docIds := make([]uint, 0)
	for docID, tokenFreqs := range normalizedFreqs {
		for token, freq := range tokenFreqs {
			docScore[docID] += freq * inverseTokenFreq[token]
		}
		docIds = append(docIds, docID)
	}

	sort.Slice(docIds, func(i, j int) bool {
		return docScore[docIds[i]] > docScore[docIds[j]]
	})

	elapsed := time.Since(start)
	log.Printf("rank docs took %s", elapsed)
	if len(docIds) <= 20 {
		return docIds
	}
	return docIds[:20]
}
