package main

import (
	"testing"
)

func TestStopWords(t *testing.T) {
	uncleanText := "HellO this is a TeSt"
	expectedTokens := map[string]int{"hello": 1, "test": 1}
	tokens := processText(uncleanText, false)

	for _, token := range tokens {
		if expectedTokens[token] == 0 {
			t.Fatalf("Unexpected token: %s\n", token)
		}
	}

}

func TestLemming(t *testing.T) {
	uncleanText := "going go went"
	expectedTokens := map[string]int{"go": 1, "went": 1}
	tokens := processText(uncleanText, true)

	for _, token := range tokens {
		if expectedTokens[token] == 0 {
			t.Fatalf("Unexpected token: %s\n", token)
		}
	}

}
