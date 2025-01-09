package main

import (
	"fmt"
)

func main() {
	index := Index{}
	for i := 0; i < 2; i++ {
		doc, err := fetchStory(i)
		if err != nil {
			continue
		}
		fmt.Printf("title: %s\n", doc.Story.Title)
		for _, comment := range doc.Comments {
			fmt.Printf("->%s\n", comment.Text)
		}
		index.add(doc)
		fmt.Printf("\n\n*******************************************************\n\n")
	}
}
