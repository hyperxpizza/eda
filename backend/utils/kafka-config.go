package utils

import (
	"bufio"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func ReadConfig(path string) (kafka.ConfigMap, error) {
	m := make(map[string]kafka.ConfigValue)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) > 0 {
			kv := strings.Split(line, "=")
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			m[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return m, nil
}
