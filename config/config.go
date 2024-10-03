package config

import (
	yaml "gopkg.in/yaml.v3"
	"os"
	"strconv"

	"testing_trainer/utils/grpc_server"
	"testing_trainer/utils/http_server"
)

const pathToConfig = "etc/config.yaml"

type Config struct {
	HTTP http_server.Config `yaml:"http"`
	GRPC grpc_server.Config `yaml:"grpc"`
	PG   Postgres           `yaml:"postgres"`
}

type Postgres struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

func (p *Postgres) GetConnectionString() string {
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + strconv.Itoa(p.Port) + "/" + p.Database + "?" + "sslmode" + "=" + p.SSLMode
}

var ConfigStruct Config

func Init() error {
	rawYaml, err := os.ReadFile(pathToConfig)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(rawYaml, &ConfigStruct); err != nil {
		return err
	}
	return nil
}
