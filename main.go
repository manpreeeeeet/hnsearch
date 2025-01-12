package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := getConnection(DBCreds{
		User:     "testuser",
		Password: "testpassword",
		DBName:   "testdb",
		Host:     "localhost",
		Port:     "5432",
	})
	if err != nil {
		fmt.Printf("ooo husbant there is an error now we are homeress %s\n", err)
	}
	err = db.AutoMigrate(&TokenModel{}, &DocumentModel{}, &CommentModel{}, &DocumentTokenFrequencyModel{})
	if err != nil {
		panic("Failed to migrate database")
	}

	index, _ := loadTokensToIndex(db)

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5174"}
	r.Use(cors.New(config))

	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		documentModels := index.searchQuery(db, query)
		documents := make([]Document, 0)
		for _, documentModel := range documentModels {
			documents = append(documents, documentModel.toDocument())
		}
		c.JSON(200, documents)
	})

	go func() {
		return
		for i := uint(42663753); i > 42660000; i-- {

			var documentModel DocumentModel
			err = db.First(&documentModel, i).Error
			if err == nil {
				continue
			}

			doc, err := fetchStory(i)
			if err != nil {
				continue
			}
			fmt.Printf("title: %s\n", doc.Story.Title)

			if err := addDocumentToDbIndex(db, doc); err != nil {
				fmt.Printf("failed to add doc to db index titled: %s\n", doc.Story.Title)
				continue
			}

			fmt.Printf("\n\n*******************************************************\n\n")
		}
	}()
	r.Run("localhost:8080")
}
