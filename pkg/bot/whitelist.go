package bot

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type WhiteList struct {
	path  string
	names map[string]struct{}
	lock  sync.Mutex
}

func LoadWhiteList(path string) (*WhiteList, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %s: %v", path, err)
	}
	defer file.Close()

	list := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		list[scanner.Text()] = struct{}{}
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	instance := &WhiteList{
		path:  path,
		names: list,
	}

	return instance, nil
}

func (whitelist *WhiteList) Write() error {
	whitelist.lock.Lock()
	defer whitelist.lock.Unlock()

	return whitelist.write()
}

func (whitelist *WhiteList) write() error {
	file, err := os.Create(whitelist.path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for line := range whitelist.names {
		_, err = fmt.Fprintln(w, line)
		if err != nil {
			return err
		}
	}
	err = w.Flush()
	return err
}

func (whitelist *WhiteList) Add(name string) error {
	whitelist.lock.Lock()
	defer whitelist.lock.Unlock()

	whitelist.names[name] = struct{}{}

	return whitelist.write()
}

func (whitelist *WhiteList) Contains(name string) bool {
	_, exists := whitelist.names[name]

	return exists
}

func (whitelist *WhiteList) Remove(name string) error {
	whitelist.lock.Lock()
	defer whitelist.lock.Unlock()

	_, exists := whitelist.names[name]

	if !exists {
		return nil
	}

	delete(whitelist.names, name)

	return whitelist.write()
}
