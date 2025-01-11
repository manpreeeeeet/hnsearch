package main

import (
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	db, _ := getConnection()
	index, _ := loadTokensToMap(db)

	docs := index.searchQuery(db, "service")
	for _, doc := range docs {
		fmt.Printf("%s\n", doc.Title)
		for _, comment := range doc.Comments {
			fmt.Printf("\n-------> %s\n", comment.Text)
		}
		fmt.Printf("\n")
	}

}
