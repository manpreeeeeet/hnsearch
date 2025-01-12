package main

import (
	"fmt"
	"testing"
)

func TestApi(t *testing.T) {
	doc, _ := fetchStory(42666572)
	fmt.Printf("%v", doc)
}
