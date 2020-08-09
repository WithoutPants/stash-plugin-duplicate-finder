package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
	DBFilename string `yaml:"db_filename"`
	Threshold  int    `yaml:"threshold"`
	AddTagName string `yaml:"add_tag_name"`
	AddDetails bool   `yaml:"add_details"`
	NewOnly    bool   `yaml:"new_only"`
}

func readConfig(fn string) (*config, error) {
	ret := &config{
		DBFilename: "df-hashstore.db",
		Threshold:  50,
	}

	_, err := os.Stat(fn)
	if err != nil {
		if os.IsNotExist(err) {
			// just return default config
			return ret, nil
		}

		return nil, err
	}

	file, err := os.Open(fn)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	parser := yaml.NewDecoder(file)
	parser.SetStrict(true)
	err = parser.Decode(&ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
