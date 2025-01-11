package main

import (
	"fmt"
	"gorm.io/gorm"
)

func (index Index) searchQuery(db *gorm.DB, query string) []DocumentModel {
	tokens := processText(query, true)
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
