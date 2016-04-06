package gaeapp

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var config struct {
	token string
}

// Load the configuration information from the config file.
func init() {
	data, err := ioutil.ReadFile("./snoreslacks.yaml")
	if err != nil {
		panic("while loading config file: " + err.Error())
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic("while loading config file: " + err.Error())
	}
}
