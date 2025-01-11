package main

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func main() {
	index := Index{}
	db, err := getConnection()
	if err != nil {
		fmt.Printf("ooo husbant there is an error now we are homeress %s\n", err)
	}
	err = db.AutoMigrate(&TokenModel{}, &DocumentModel{}, &CommentModel{})
	if err != nil {
		panic("Failed to migrate database")
	}

	for i := uint(42667962); i > 42660000; i-- {
		doc, err := fetchStory(i)
		if err != nil {
			continue
		}
		fmt.Printf("title: %s\n", doc.Story.Title)
		index.add(doc)

		var documentModel DocumentModel
		err = db.First(&documentModel, i).Error
		if err != nil {
			documentModel = *doc.toDocumentModel()
			db.Create(&documentModel)
		}

		tokens := doc.getTokens()
		tokensModel := make([]TokenModel, 0)
		for _, token := range tokens {
			var tokenModel TokenModel
			err := db.Where("token = ?", token).First(&tokenModel).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					tokenModel = TokenModel{Token: token}
					if createErr := db.Create(&tokenModel).Error; createErr != nil {

					}

				} else {

				}
			}

			tokensModel = append(tokensModel, tokenModel)
		}
		documentModel.Tokens = tokensModel
		if err := db.Save(documentModel).Error; err != nil {
			fmt.Printf("failed to associate terms with document: %s", err)
		}

		fmt.Printf("\n\n*******************************************************\n\n")
	}
}
