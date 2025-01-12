package main

import (
	"fmt"
	"github.com/bbalet/stopwords"
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

func TestParsing(t *testing.T) {
	uncleanText := "I would like to suggest, as a hypothetical, that well funded companies (you know who) are promoting this AI &quot;danger&quot; as a way to cajole federal legislation to come out they can &quot;help with&quot;.<p>It is the fear of every U.S. company to have to deal with 50 different state&#x27;s legislation (see guns, porn, abortion, alcohol, gambling, prostitution, etc. that all vary state by state) because it&#x27;s more expensive to have to bend to every little state&#x27;s ideas of how to protect their citizens.<p>They want a single point at which they can lobby (and lobby <i>hard</i> with <i>lots of money</i>) to get the lax, self-dealing laws with lots of legal protection they so desperately know will provide them with super-profits.<p>Of course, I could be wrong. Maybe Skynet is just around the corner."
	stop := stopwords.CleanString(uncleanText, "en", true)
	fmt.Printf("stop %v", stop)
}
