package config

import (
	"bufio"
	"fmt"
	"os"
)

func ReadWhitelist(path string) (list []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	err = scanner.Err()
	return
}

func WriteWhitelist(path string, lines []string) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		_, err = fmt.Fprintln(w, line)
		if err != nil {
			return
		}
	}
	err = w.Flush()
	return
}
