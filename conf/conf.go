package conf

import (
	"cloud/log"
	"gopkg.in/yaml.v3"
	"os"
)

var (
	MongoConfig Mongo
	ServerConfig Server
)

type Config struct {
	Mongo `yaml:"mongo"`
	Server `yaml:"server"`
}

type Mongo struct {
	Uri string `yaml:"uri"`
	DB string `yaml:"db"`
}

type Server struct {
	IP string `yaml:"ip"`
	Port uint32 `yaml:"port"`
}

func init() {
	yfile, err := os.Open("conf/conf.yaml")
	defer yfile.Close()
	if err != nil {
		log.Panic("err in conf Open yaml,err=%v", err)
	}
	ydecode := yaml.NewDecoder(yfile)

	var c Config
	err = ydecode.Decode(&c)
	if err != nil {
		log.Panic("err in conf Decode,err=%v", err)
	}

	MongoConfig = c.Mongo
	ServerConfig = c.Server
}