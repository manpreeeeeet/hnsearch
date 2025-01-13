package main

import (
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	db, _ := getConnection(DBCreds{
		User:     "testuser",
		Password: "testpassword",
		DBName:   "testdb",
		Host:     "localhost",
		Port:     "5432",
	})
	index, _ := loadTokensToIndex(db)

	for i := 0; i < 10; i++ {

		docs := index.searchQuery(db, "ai art")
		for _, doc := range docs {
			fmt.Printf("%s\n", doc.Title)
			//for _, comment := range doc.Comments {
			//	fmt.Printf("\n-------> %s\n", comment.Text)
			//}
			fmt.Printf("\n")
		}
	}

}
