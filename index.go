package main

import (
	"github.com/bbalet/stopwords"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/reiver/go-porterstemmer"
	html2 "html"
	"strings"
)

type Index map[string][]uint

// normalize -> tokenize -> remove stop words -> stemming
func processText(text string, stem bool) []string {
	clean := stopwords.CleanString(text, "en", true)
	tokens := strings.Fields(clean)
	if stem {
		for i, _ := range tokens {
			tokens[i] = porterstemmer.StemString(tokens[i])
		}
	}

	return tokens
}

func safeRemoveHtml(text string) string {
	p := bluemonday.UGCPolicy()
	html := p.Sanitize(text)
	stripped := strip.StripTags(html)
	return html2.UnescapeString(stripped)
}

func (index Index) add(doc *Document) {
	titleTokens := processText(doc.Story.Title, true)
	for _, token := range titleTokens {
		index[token] = append(index[token], doc.Id)
	}

	for _, comment := range doc.Comments {
		commentTokens := processText(comment.Text, true)
		for _, token := range commentTokens {
			index[token] = append(index[token], doc.Id)
		}
	}
}

func (doc *Document) getTokens() (tokens map[string]int) {
	tokens = make(map[string]int)
	titleTokens := processText(doc.Story.Title, true)
	for _, token := range titleTokens {
		tokens[token] = tokens[token] + 1
	}

	for _, comment := range doc.Comments {
		commentTokens := processText(comment.Text, true)
		for _, token := range commentTokens {
			tokens[token] = tokens[token] + 1
		}
	}
	return tokens
}
