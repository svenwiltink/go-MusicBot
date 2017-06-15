package player

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"sync"
	"time"
)

type StatsStorage struct {
	stats *Statistics
	path  string
	mutex sync.Mutex
}

func NewStatsStorage(path string) (qs *StatsStorage) {
	return &StatsStorage{
		path: path,
	}
}

func (s *StatsStorage) OnStatsUpdate(args ...interface{}) {
	if len(args) < 1 {
		return
	}
	stats, ok := args[0].(*Statistics)
	if !ok {
		return
	}

	err := s.SaveStats(stats)
	if err != nil {
		logrus.Errorf("StatsStorage.OnStatsUpdate: Error saving stats: %v", err)
		return
	}
}

func (s *StatsStorage) SaveStats(stats *Statistics) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	buf, err := json.Marshal(stats)
	if err != nil {
		logrus.Errorf("StatsStorage.SaveStats: Error marshalling json: %v", err)
		return
	}

	err = ioutil.WriteFile(s.path, buf, 0755)
	if err != nil {
		logrus.Errorf("StatsStorage.SaveStats: Error saving file [%s] %v", s.path, err)
		return
	}
	return
}

func (s *StatsStorage) ReadStats() (stats *Statistics, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	buf, err := ioutil.ReadFile(s.path)
	if err != nil {
		logrus.Warnf("StatsStorage.ReadStats: Error reading file [%s] %v", s.path, err)
		return
	}

	stats = &Statistics{
		TimeByPlayer:        make(map[string]time.Duration),
		SongsPlayedByPlayer: make(map[string]int),
		SongsAddedByUser:    make(map[string]int),
	}
	err = json.Unmarshal(buf, stats)
	if err != nil {
		logrus.Errorf("StatsStorage.ReadStats: Error unmarshalling json: %v", err)
		return
	}
	return
}
