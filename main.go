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
		fmt.Printf("error connecting to db%s\n", err)
	}
	err = db.AutoMigrate(&TokenModel{}, &DocumentModel{}, &CommentModel{}, &DocumentTokenFrequencyModel{}, &CommentTokenFrequencyModel{}, &ResolvedItemModel{})
	if err != nil {
		panic("Failed to migrate database")
	}

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	r.Use(cors.New(config))

	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		documentModels := searchQuery(db, query)
		documents := make([]Document, 0)
		for _, documentModel := range documentModels {
			documents = append(documents, documentModel.toDocument())
		}
		c.JSON(200, documents)
	})

	go func() {
		resumeHnIndexing(db, true, 1000)
	}()

	r.Run("localhost:8080")
}
