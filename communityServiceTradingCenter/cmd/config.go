package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

type rootConfig struct {
	LogDir         string `json:"logDir"`
	DbPath         string `json:"dbPath"`
	StaticServeDir string `json:"staticServeDir"`
	ServeAddress   string `json:"serveAddress"`
}

var defaultConfig = rootConfig{
	LogDir:         "./log",
	DbPath:         "./db",
	StaticServeDir: "./static",
	ServeAddress:   ":12345",
}

func readConfig() rootConfig {
	var err error
	defer func() {
		if err != nil {
			saveConfig(defaultConfig)
		}
	}()
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("read config failed : " + err.Error() + "   using default config and save it to local disk")
		return defaultConfig
	}
	var res rootConfig
	err = json.Unmarshal(data, &res)
	if err != nil {
		fmt.Println("parse config failed : " + err.Error() + "   using default config and save it to local disk")
		return defaultConfig
	}

	return res

}

func saveConfig(rc rootConfig) {
	data, _ := json.MarshalIndent(rc, "", "  ")
	_ = os.WriteFile("config.json", data, 0777)
}
