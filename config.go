package main

import (
	"encoding/json"
	"os"
)

const DEFAULT_FILENAME = "conf.json"

type Config struct {
	File      string
	Domain    string
	Host      string
	Command   string
	BackupDir string
	Addresses []string
}

// Reads config from file
func ReadConfig(filename string) Result[Config] {
	file, _ := os.Open(filename)
	defer file.Close()

	var config Config
	err := json.NewDecoder(file).Decode(&config)
	if err != nil {
		return ResultErr[Config](err)
		// log.Fatal("Error reading config: ", err.Error())
	}
	return ResultOk(config)
}

// Reads config from default place
func ReadConfigDefault() Result[Config] {
	return ReadConfig(DEFAULT_FILENAME)
}
