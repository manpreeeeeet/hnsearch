package main

import (
	"github.com/bbalet/stopwords"
	"github.com/reiver/go-porterstemmer"
	"strings"
)

type Index map[string][]int

// normalize -> tokenize -> remove stop words -> stemming
func processText(text string, stem bool) []string {
	clean := stopwords.CleanString(text, "en", false)
	tokens := strings.Fields(clean)
	if stem {
		for i, _ := range tokens {
			tokens[i] = porterstemmer.StemString(tokens[i])
		}
	}

	return tokens
}

func (index Index) add(doc *Document) {
	titleTokens := processText(doc.Story.Title, false)
	for _, token := range titleTokens {
		index[token] = append(index[token], doc.Id)
	}

	for _, comment := range doc.Comments {
		commentTokens := processText(comment.Text, false)
		for _, token := range commentTokens {
			index[token] = append(index[token], doc.Id)
		}
	}
}
