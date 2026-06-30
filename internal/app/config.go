package app

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port              string
	PublicBaseURL     string
	WrapperAPIKey     string
	AdminToken        string
	BananaBaseURL     string
	BananaAPIKey      string
	ModelPrefix       string
	MaxWorkers        int
	MaxQueue          int
	PollInterval      time.Duration
	RequestTimeout    time.Duration
	BananaHTTPTimeout time.Duration
	ReturnB64JSON     bool
}

func LoadConfig() Config {
	return Config{
		Port:              env("PORT", "3000"),
		PublicBaseURL:     strings.TrimRight(env("PUBLIC_BASE_URL", ""), "/"),
		WrapperAPIKey:     env("WRAPPER_API_KEY", ""),
		AdminToken:        env("ADMIN_TOKEN", ""),
		BananaBaseURL:     strings.TrimRight(env("BANANA_API_BASE", "https://nb.gettoken.cn/openapi"), "/"),
		BananaAPIKey:      env("BANANA_API_KEY", ""),
		ModelPrefix:       env("MODEL_PREFIX", "banana-pro"),
		MaxWorkers:        envInt("MAX_WORKERS", 512),
		MaxQueue:          envInt("MAX_QUEUE", 20000),
		PollInterval:      time.Duration(envInt("BANANA_POLL_INTERVAL_MS", 2500)) * time.Millisecond,
		RequestTimeout:    time.Duration(envInt("REQUEST_TIMEOUT_SECONDS", 600)) * time.Second,
		BananaHTTPTimeout: time.Duration(envInt("BANANA_HTTP_TIMEOUT_SECONDS", 60)) * time.Second,
		ReturnB64JSON:     envBool("RETURN_B64_JSON", false),
	}
}

func (cfg Config) Models() []string {
	return []string{cfg.ModelPrefix + "-1k", cfg.ModelPrefix + "-2k", cfg.ModelPrefix + "-4k"}
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
