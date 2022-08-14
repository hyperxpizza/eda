package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type PostgresConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Host     string `json:"host"`
}

func NewConfig(pathToFile string) (*PostgresConfig, error) {
	file, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var c PostgresConfig

	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
