package conf

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port  string `yaml:"port"`
		Url   string `yaml:"url"`
		Route string    `yaml:"route"`
		DataLink string `yaml:"datalink"`
		PrivateLink string `yaml:"privatelink"`
	} `yaml:"server"`

	Mongo struct {
		Db         string `yaml:"db"`
		Collection string `yaml:"collection"`
		Url        string `yaml:"url"`
		Timeout    int    `yaml:"timeout"`
		Field      string `yaml:"field"`
	} `yaml:"mongodb"`
}



// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}


