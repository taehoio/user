package config

import (
	"log"
	"os"
	"strconv"
)

type Setting struct {
	ServiceName        string
	GRPCServerEndpoint string
	GRPCServerPort     string
	HTTPServerPort     string

	MysqlNetworkType  string
	MysqlAddress      string
	MysqlUser         string
	MysqlPassword     string
	MysqlDatabaseName string

	Env                       string
	GracefulShutdownTimeoutMs int
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if defaultValue == "" {
		log.Fatalf("a required environment variable missed: %s", key)
	}
	return defaultValue
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Panic(err)
	}
	return i
}

func NewSetting() Setting {
	return Setting{
		ServiceName:        "user",
		GRPCServerEndpoint: getEnv("GRPC_SERVER_ENDPOINT", "localhost:18081"),
		GRPCServerPort:     getEnv("GRPC_SERVER_PORT", "18081"),
		HTTPServerPort:     getEnv("HTTP_SERVER_PORT", "18082"),

		MysqlNetworkType:  getEnv("MYSQL_NETWORK_TYPE", "tcp"),
		MysqlAddress:      getEnv("MYSQL_ADDRESS", "localhost"),
		MysqlUser:         getEnv("MYSQL_USER", "taehoio_sa"),
		MysqlPassword:     getEnv("MYSQL_PASSWORD", ""),
		MysqlDatabaseName: getEnv("MYSQL_DATABASE_NAME", "taehoio"),

		Env:                       getEnv("ENV", "development"),
		GracefulShutdownTimeoutMs: mustAtoi(getEnv("GRACEFUL_SHUTDOWN_TIMEOUT_MS", "5000")),
	}
}
