package bot

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type AllowList struct {
	path  string
	names map[string]struct{}
	lock  sync.Mutex
}

func LoadAllowList(path string) (*AllowList, error) {
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

	instance := &AllowList{
		path:  path,
		names: list,
	}

	return instance, nil
}

func (allowlist *AllowList) Write() error {
	allowlist.lock.Lock()
	defer allowlist.lock.Unlock()

	return allowlist.write()
}

func (allowlist *AllowList) write() error {
	file, err := os.Create(allowlist.path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for line := range allowlist.names {
		_, err = fmt.Fprintln(w, line)
		if err != nil {
			return err
		}
	}
	err = w.Flush()
	return err
}

func (allowlist *AllowList) Add(name string) error {
	allowlist.lock.Lock()
	defer allowlist.lock.Unlock()

	allowlist.names[name] = struct{}{}

	return allowlist.write()
}

func (allowlist *AllowList) Contains(name string) bool {
	_, exists := allowlist.names[name]

	return exists
}

func (allowlist *AllowList) Remove(name string) error {
	allowlist.lock.Lock()
	defer allowlist.lock.Unlock()

	_, exists := allowlist.names[name]

	if !exists {
		return nil
	}

	delete(allowlist.names, name)

	return allowlist.write()
}
