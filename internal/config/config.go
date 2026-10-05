package config

import "os"

type Config struct {
	InputDir  string
	OutputDir string
}

func Load() Config {
	return Config{
		InputDir:  getenv("INPUT_DIR", "./input"),
		OutputDir: getenv("OUTPUT_DIR", "./output"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
