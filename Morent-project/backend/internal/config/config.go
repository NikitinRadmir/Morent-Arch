package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	MinioEndpoint   string `json:"minio_endpoint"`
	MinioAccessKey  string `json:"minio_access_key"`
	MinioSecretKey  string `json:"minio_secret_key"`
	MinioBucket     string `json:"minio_bucket"`
	MinioUseSSL     bool   `json:"minio_use_ssl"`

	GRPCPort string `json:"grpc_port"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	file, errOpen := os.Open(path)
	if errOpen != nil {
		return nil, errOpen
	}
	defer file.Close()
	errDecode := json.NewDecoder(file).Decode(&cfg)
	return &cfg, errDecode
}


