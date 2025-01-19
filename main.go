package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		args = append(args, "dev")
	}

	if args[1] == "dev" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	db, err := getConnection(DBCreds{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	})
	if err != nil {
		fmt.Printf("error connecting to db%s\n", err)
	}
	err = db.AutoMigrate(&TokenModel{}, &DocumentModel{}, &CommentModel{}, &DocumentTokenFrequencyModel{}, &CommentTokenFrequencyModel{}, &ResolvedItemModel{})
	//query := db.Model(&DocumentTokenFrequencyModel{}).Select("document_id, SUM(frequency) as total_tokens").Group("document_id")
	//err = db.Migrator().CreateView("document_token_counts_view", gorm.ViewOption{Query: query})
	//if err != nil {
	//	return
	//}

	if err != nil {
		panic("Failed to migrate database")
	}

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:8080", "http://hnsearch.mnprt.me"}
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

	if os.Getenv("START_INDEX") == "true" {
		maxItem, _ := strconv.Atoi(os.Getenv("MAX_ITEMS"))
		go func() {
			resumeHnIndexing(db, true, int64(maxItem))
		}()
	}
	r.Run("0.0.0.0:8081")
}
