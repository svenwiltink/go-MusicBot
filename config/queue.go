package config

import (
	"bufio"
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/songplayer"
	"os"
	"strings"
)

type QueueStorage struct {
	path string
}

func NewQueueStorage(path string) (qs *QueueStorage) {
	return &QueueStorage{
		path: path,
	}
}

func (qs *QueueStorage) OnListUpdate(args ...interface{}) {
	if len(args) < 1 {
		return
	}
	queue, ok := args[0].([]songplayer.Playable)
	if !ok {
		return
	}
	var urls []string
	for _, song := range queue {
		urls = append(urls, song.GetURL())
	}

	err := qs.saveQueue(urls)
	if err != nil {
		fmt.Printf("[QueueStorage] Error saving queue: %v", err)
	}
}

func (qs *QueueStorage) saveQueue(urls []string) (err error) {
	file, err := os.Create(qs.path)
	if err != nil {
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, url := range urls {
		_, err = fmt.Fprintln(w, url)
		if err != nil {
			return
		}
	}
	err = w.Flush()
	return
}

func (qs *QueueStorage) ReadQueue() (urls []string, err error) {
	file, err := os.Open(qs.path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, strings.TrimSpace(scanner.Text()))
	}
	err = scanner.Err()
	return
}
