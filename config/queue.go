package config

import (
	"bufio"
	"fmt"
	"github.com/SvenWiltink/go-MusicBot/songplayer"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

type QueueStorage struct {
	path  string
	mutex sync.Mutex
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
		logrus.Warnf("QueueStorage.OnListUpdate: Error saving queue: %v", err)
		return
	}
}

func (qs *QueueStorage) saveQueue(urls []string) (err error) {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	file, err := os.Create(qs.path)
	if err != nil {
		logrus.Warnf("QueueStorage.saveQueue: Error opening file [%s] %v", qs.path, err)
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
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	file, err := os.Open(qs.path)
	if err != nil {
		logrus.Warnf("QueueStorage.ReadQueue: Error reading file [%s] %v", qs.path, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, strings.TrimSpace(scanner.Text()))
	}
	err = scanner.Err()
	if err != nil {
		logrus.Warnf("QueueStorage.ReadQueue: Error scanning file: %v", err)
		return
	}
	return
}
