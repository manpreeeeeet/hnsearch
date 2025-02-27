package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

const baseURL = "https://hacker-news.firebaseio.com/v0"

type ItemType string

const (
	StoryType   ItemType = "story"
	CommentType ItemType = "comment"
)

type Item struct {
	ID      uint     `json:"id"`
	Type    ItemType `json:"type"`
	By      string   `json:"by"`
	Time    int64    `json:"time"`
	Deleted bool     `json:"deleted"`
	Dead    bool     `json:"dead"`
}

type Story struct {
	Item
	Title string `json:"title"`
	URL   string `json:"url"`
	Score int    `json:"score"`
	Kids  []uint `json:"kids"`
}

type Comment struct {
	Item
	Parent   int       `json:"parent"`
	Text     string    `json:"text"`
	Kids     []uint    `json:"kids"`
	Comments []Comment `json:"comments"`
}

type Document struct {
	Id       uint      `json:"id"`
	Story    Story     `json:"story"`
	Comments []Comment `json:"comments"`
}

func fetchLatest() (int, error) {
	url := fmt.Sprintf("%s/maxitem.json", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, err
	}
	return id, nil
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
	storyChannel := make(chan interface{})
	storyWaitGroup.Add(1)
	go fetchItem(id, &storyWaitGroup, storyChannel)
	go func() {
		storyWaitGroup.Wait()
		close(storyChannel)
	}()

	item := <-storyChannel

	switch v := item.(type) {
	case Story:
		if v.Score <= 5 {
			log.Printf("Debug: ignoring story %d since score <= 1", v.ID)
			return nil, fmt.Errorf("id: %d score too low\n", id)
		}
		doc := Document{
			Id:       id,
			Story:    item.(Story),
			Comments: make([]Comment, 0),
		}
		if v.Dead || v.Deleted {
			log.Printf("Debug: ignoring story comments since story: %d is either dead or deleted.\n", v.ID)
			return &doc, nil
		}

		if doc.Story.URL == "" {
			doc.Story.URL = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", doc.Story.ID)
		}

		commentChannel := make(chan interface{})
		commentWaitGroup := &sync.WaitGroup{}

		for _, kid := range v.Kids {
			commentWaitGroup.Add(1)
			go fetchItem(kid, commentWaitGroup, commentChannel)
		}

		go func() {
			commentWaitGroup.Wait()
			close(commentChannel)
		}()

		for result := range commentChannel {
			comment, ok := result.(Comment)
			if !ok {
				log.Printf("%v is not a comment\n", comment)
				continue
			}
			if comment.Deleted || comment.Dead {
				log.Printf("Debug: ignoring comment: %d since it is either dead or deleted.\n", comment.ID)
				continue
			}

			doc.Comments = append(doc.Comments, comment)
		}
		return &doc, nil

	default:
		return nil, fmt.Errorf("id: %d is not a story\n", id)
	}
}
