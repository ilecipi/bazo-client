package util

import (
	"os"
	"fmt"
	"encoding/json"
)

const (
	CONFIGURATION_FILE = "configuration.json"

	HEALTH_CHECK_INTERVAL = 30 //Sec
	MIN_MINERS = 2
)

var (
	Config Configuration
)

type Configuration struct {
	ThisIpport string
	Thisclient struct {
		Ip   string `json:"ip"`
		Port string `json:"port"`
	} `json:"this_client"`
	BootstrapIpport string
	Bootstrapserver struct {
		Ip   string `json:"ip"`
		Port string `json:"port"`
	} `json:"bootstrap_server"`
}

func LoadConfiguration() (config Configuration) {
	configFile, err := os.Open(CONFIGURATION_FILE)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	config.ThisIpport = config.Thisclient.Ip + ":" + config.Thisclient.Port
	config.BootstrapIpport = config.Bootstrapserver.Ip + ":" + config.Bootstrapserver.Port
	return config
}
