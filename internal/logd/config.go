package logd

import (
	"context"
	"encoding/json"
	"log"
	"os"
)

const LOGS_DIRECTORY = "/var/log/ravel"

type LogdConfig struct {
	Exporter       string                 `json:"exporter" default:"stdout"`
	ExporterConfig map[string]interface{} `json:"exporter_config"`
}

func LoadConfig(filePath string) (*LogdConfig, error) {

	// Load the config
	var config LogdConfig
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func NewLogger(ctx context.Context) (*log.Logger, context.Context, error) {
	logFile := LOGS_DIRECTORY
	writer, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, ctx, err
	}

	defer writer.Close()

	l := log.New(writer, "", log.LstdFlags)

	ctx = context.WithValue(ctx, LoggerKey, l)

	return l, ctx, nil
}

func GetLogger(ctx context.Context) (*log.Logger, error) {
	l := ctx.Value(LoggerKey)
	if l == nil {
		return nil, nil
	}

	return l.(*log.Logger), nil
}
