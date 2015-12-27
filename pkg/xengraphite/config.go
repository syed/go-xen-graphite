package xengraphite

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

const (
	CONF_FILE = "xengraphite.json"
)

func FindConfigFile() string {

	//look in current directory
	if _, err := os.Stat(CONF_FILE); err == nil {
		return CONF_FILE
	}

	//look in /etc
	if _, err := os.Stat("/etc/" + CONF_FILE); err == nil {
		return "/etc/" + CONF_FILE
	}

	return ""
}

func ParseConfigFile(file_path string) *Config {

	conf := new(Config)
	conf_data, _ := ioutil.ReadFile(file_path)

	if err := json.Unmarshal(conf_data, &conf); err != nil {
		log.Errorf(err.Error())
	}

	return conf
}
