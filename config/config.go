package config

import (
	"os"

	"gopkg.in/yaml.v3"
	"strconv"
)

const pathToConfig = "etc/config.yaml"

type Config struct {
	HTTP Http     `yaml:"http"`
	PG   Postgres `yaml:"postgres"`
}

type Http struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
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
