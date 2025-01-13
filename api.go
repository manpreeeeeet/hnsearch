package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const baseURL = "https://hacker-news.firebaseio.com/v0"

type ItemType string

const (
	StoryType   ItemType = "story"
	CommentType ItemType = "comment"
)

type Item struct {
	ID   uint     `json:"id"`
	Type ItemType `json:"type"`
	By   string   `json:"by"`
	Time int64    `json:"time"`
}

type Story struct {
	Item
	Title string `json:"title"`
	URL   string `json:"url"`
	Score int    `json:"score"`
	Kids  []int  `json:"kids"`
}

type Comment struct {
	Item
	Parent   int       `json:"parent"`
	Text     string    `json:"text"`
	Kids     []int     `json:"kids"`
	Comments []Comment `json:"comments"`
}

type Document struct {
	Id       uint      `json:"id"`
	Story    Story     `json:"story"`
	Comments []Comment `json:"comments"`
}

func fetchItem(id uint, wg *sync.WaitGroup, ch chan<- interface{}) {
	defer wg.Done()
	url := fmt.Sprintf("%s/item/%d.json", baseURL, id)

	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Errorf("error fetching item %d: %v", id, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- fmt.Errorf("failed to fetch item %d: %s", id, resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- fmt.Errorf("error reading body for item %d: %v", id, err)
		return
	}

	var item Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		ch <- fmt.Errorf("error unmarshaling item %d: %v", id, err)
		return
	}

	switch item.Type {
	case StoryType:
		var story Story
		err = json.Unmarshal(body, &story)
		if err != nil {
			ch <- fmt.Errorf("error unmarshaling story %d: %v", id, err)
			return
		}
		story.ID = item.ID
		story.By = item.By
		ch <- story

	case CommentType:
		var comment Comment
		err = json.Unmarshal(body, &comment)
		if err != nil {
			ch <- fmt.Errorf("error unmarshaling comment %d: %v", id, err)
			return
		}
		comment.ID = item.ID
		comment.By = item.By
		ch <- comment

	default:
		ch <- item
	}
}

func fetchStory(id uint) (*Document, error) {
	var storyWaitGroup sync.WaitGroup
	ch := make(chan interface{})
	storyWaitGroup.Add(1)
	go fetchItem(id, &storyWaitGroup, ch)
	go func() {
		storyWaitGroup.Wait()
		close(ch)
	}()

	item := <-ch

	switch v := item.(type) {
	case Story:
		doc := Document{
			Id:       id,
			Story:    item.(Story),
			Comments: make([]Comment, 0),
		}
		commentChannel := make(chan interface{})
		commentWaitGroup := &sync.WaitGroup{}
		for _, kid := range v.Kids {
			commentWaitGroup.Add(1)
			go fetchCommentTree(uint(kid), commentWaitGroup, commentChannel)
		}

		go func() {
			commentWaitGroup.Wait()
			close(commentChannel)
		}()

		for result := range commentChannel {
			comment, ok := result.(Comment)
			if !ok {
				fmt.Println("not a comment")
			}
			doc.Comments = append(doc.Comments, comment)
		}
		return &doc, nil

	default:
		return nil, fmt.Errorf("id: %d is not a story\n", id)
	}
}

func fetchCommentTree(id uint, wg *sync.WaitGroup, ch chan<- interface{}) {
	defer wg.Done()

	// Fetch the current comment
	var commentWG sync.WaitGroup
	commentCh := make(chan interface{})
	commentWG.Add(1)
	go fetchItem(id, &commentWG, commentCh)

	go func() {
		commentWG.Wait()
		close(commentCh)
	}()

	// Get the comment result
	result := <-commentCh
	comment, ok := result.(Comment)
	if !ok {
		if _, isErr := result.(error); isErr {
			fmt.Printf("error fetching comment:%d", id)
		}
		return
	}

	// If the comment has children, fetch them recursively
	if len(comment.Kids) > 0 {
		childrenChannel := make(chan interface{})
		childrenWG := &sync.WaitGroup{}

		// Start fetching all child comments
		for _, kid := range comment.Kids {
			childrenWG.Add(1)
			go fetchCommentTree(uint(kid), childrenWG, childrenChannel)
		}

		// Wait for all children to be fetched
		go func() {
			childrenWG.Wait()
			close(childrenChannel)
		}()

		// Collect all child comments
		comment.Comments = make([]Comment, 0, len(comment.Kids))
		for childResult := range childrenChannel {
			if childComment, ok := childResult.(Comment); ok {
				comment.Comments = append(comment.Comments, childComment)
			} else if err, ok := childResult.(error); ok {
				fmt.Printf("Error fetching child comment: %v\n", err)
			}
		}
	}

	// Send the complete comment with all its children
	ch <- comment
}
