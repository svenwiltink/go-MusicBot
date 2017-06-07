package config

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

func ReadWhitelist(path string) (list []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		logrus.Errorf("config.ReadWhitelist: Error opening file: [%s] %v", path, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		logrus.Errorf("config.ReadWhitelist: Error scanning file: [%s] %v", path, err)
		return
	}
	return
}

func WriteWhitelist(path string, lines []string) (err error) {
	file, err := os.Create(path)
	if err != nil {
		logrus.Errorf("config.WriteWhitelist: Error opening file: [%s] %v", path, err)
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		_, err = fmt.Fprintln(w, line)
		if err != nil {
			logrus.Errorf("config.WriteWhitelist: Error writing: [%s] %v", path, err)
			return
		}
	}
	err = w.Flush()
	if err != nil {
		logrus.Errorf("config.WriteWhitelist: Error flushing writer: [%s] %v", path, err)
		return
	}
	return
}
