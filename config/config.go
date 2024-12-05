package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"

	"gopkg.in/yaml.v3"
	"strconv"
)

const (
	pathToConfig  = "./etc/config.yaml"
	pathToEnvFile = "./etc/config.env"
)

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

func InitConfigWithEnvs() error {
	var envs map[string]string

	err := godotenv.Load(pathToEnvFile)
	if err != nil {
		return fmt.Errorf("godotenv.Load: %w", err)
	}
	envs, err = godotenv.Read(pathToEnvFile)
	if err != nil {
		return fmt.Errorf("godotenv.Load: %w", err)
	}

	pgHost := envs["PG_HOST"]
	pgPort, err := strconv.Atoi(envs["PG_PORT"])
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}
	pgUser := envs["PG_USER"]
	pgPassword := envs["PG_PASSWORD"]
	pgDatabase := envs["PG_DATABASE"]
	pgSSLMode := envs["PG_SSLMODE"]

	httpHost := envs["HTTP_HOST"]
	httpPort, err := strconv.Atoi(envs["HTTP_PORT"])
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}

	ConfigStruct = Config{
		HTTP: Http{
			Host: httpHost,
			Port: httpPort,
		},
		PG: Postgres{
			Host:     pgHost,
			Port:     pgPort,
			User:     pgUser,
			Password: pgPassword,
			Database: pgDatabase,
			SSLMode:  pgSSLMode,
		},
	}
	return nil
}
