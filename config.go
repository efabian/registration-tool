package main

import (
	"log"
	"os"
)

// ZoomConfig holds Zoom meeting details for a specific day.
type ZoomConfig struct {
	Link string
	Meet string
	Pass string
	Date string
	QR   string
}

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL   string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	EmailSender   string
	EmailBCC      string
	EmailBCCNick  string
	EmailSubject  string
	AdminPasscode string
	// PingSalt is used by PingHandler to produce an uptime token.
	// FIX C-3: was previously a hardcoded constant in handlers.go.
	PingSalt string
	Zoom     map[string]ZoomConfig
}

// cfg is the package-level configuration instance.
var cfg Config

// loadConfig reads all required environment variables and populates cfg.
// It calls log.Fatal if any required variable is missing.
func loadConfig() {
	cfg = Config{
		DatabaseURL:   requireEnv("DATABASE_URL"),
		SMTPHost:      requireEnv("SMTP_HOST"),
		SMTPPort:      getEnvDefault("SMTP_PORT", "587"),
		SMTPUser:      requireEnv("SMTP_USER"),
		SMTPPass:      requireEnv("SMTP_PASS"),
		EmailSender:   requireEnv("EMAIL_SENDER"),
		EmailBCC:      requireEnv("EMAIL_BCC"),
		EmailBCCNick:  requireEnv("EMAIL_BCC_NICK"),
		EmailSubject:  requireEnv("EMAIL_SUBJECT"),
		AdminPasscode: requireEnv("ADMIN_PASSCODE"),
		PingSalt:      requireEnv("PING_SALT"),
		Zoom:          loadZoomConfig(),
	}
}

// loadZoomConfig reads all Zoom-related env vars and returns a map keyed by day abbreviation.
func loadZoomConfig() map[string]ZoomConfig {
	days := []string{"MON", "TUE", "WED", "THU", "FRI", "SAT"}
	zoom := make(map[string]ZoomConfig, len(days))
	for _, d := range days {
		key := d // e.g. "MON"
		zoom[key] = ZoomConfig{
			Link: requireEnv("ZOOM_" + key + "_LINK"),
			Meet: requireEnv("ZOOM_" + key + "_MEET"),
			Pass: requireEnv("ZOOM_" + key + "_PASS"),
			Date: requireEnv("ZOOM_" + key + "_DATE"),
			QR:   requireEnv("ZOOM_" + key + "_QR"),
		}
	}
	return zoom
}

// requireEnv returns the value of an environment variable or calls log.Fatal if it is not set.
func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}

// getEnvDefault returns the value of an environment variable or a default value if not set.
func getEnvDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
