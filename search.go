package main

import (
	"fmt"
	"gorm.io/gorm"
	"sort"
)

func (index Index) searchQuery(db *gorm.DB, query string) []DocumentModel {
	tokens := processText(query, true)
	documents := index.searchDocuments(db, tokens)
	documents = rankDocuments(db, tokens, documents)
	return documents
}

func (index Index) searchDocuments(db *gorm.DB, tokens []string) []DocumentModel {
	documentIds := map[uint]uint{}

	for _, token := range tokens {
		for _, documentId := range index[token] {
			documentIds[documentId] = 1
		}
	}

	documents := make([]DocumentModel, 0)
	for documentId := range documentIds {
		var documentModel DocumentModel
		err := db.First(&documentModel, documentId).Error
		if err != nil {
			continue
		}
		var comments []CommentModel
		err = db.Model(&documentModel).Association("Comments").Find(&comments)
		if err != nil {
			fmt.Printf("oooo watttt no comments????")
		}
		documentModel.Comments = comments
		documents = append(documents, documentModel)
	}

	return documents

}

func rankDocuments(db *gorm.DB, tokens []string, docs []DocumentModel) []DocumentModel {
	// find frequency of tokens in each document
	// ivf (how rare this thing is) -> count(id) where token = ?
	tokenFreqWeight := 0.1
	inverseFreqWeight := 0.2

	inverseTokenFreq := map[string]int64{}
	docScore := map[uint]float64{}
	for _, token := range tokens {

		tokenModel := getTokenModel(db, token)
		if tokenModel == nil {
			continue
		}

		inverseTokenFreq[token] = getInverseDocumentFrequency(db, tokenModel.ID)
		for _, doc := range docs {
			tokenFreq := getTokenFrequency(db, doc.ID, tokenModel.ID)
			docScore[doc.ID] += (tokenFreqWeight * float64(tokenFreq)) + (inverseFreqWeight * float64(inverseTokenFreq[token]))
		}
	}

	sort.Slice(docs, func(i, j int) bool {
		return docScore[docs[i].ID] > docScore[docs[j].ID]
	})
	return docs
}
