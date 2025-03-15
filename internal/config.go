package internal

import (
	"encoding/json"
	"log/slog"
	"os"
)

type Discord struct {
	Token     string  `json:"token"`
	GuildId   string  `json:"guild"`
	ChannelId string  `json:"channel_id"`
	MessageId *string `json:"message_id"`
}

type Config struct {
	Discord             *Discord `json:"discord"`
	Servers             []Server `json:"servers"`
	PollIntervalSeconds *int     `json:"poll_interval_seconds"`

	path string
}

type Server struct {
	Name        string      `json:"name"`
	Color       *int        `json:"color"`
	ServiceId   string      `json:"service_id"`
	Credentials Credentials `json:"credentials"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Config) Save() error {
	config, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, config, 0655)
}

func NewConfig(path string, logger *slog.Logger) (*Config, error) {
	config, err := readConfig(path, logger)
	if err != nil {
		return config, err
	}

	return config, config.Save()
}

func readConfig(path string, logger *slog.Logger) (*Config, error) {
	var config Config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Info("create-config")
		config = Config{}
	} else {
		logger.Info("read-existing-config")
		c, err := os.ReadFile(path)
		if err != nil {
			return &Config{}, err
		}
		err = json.Unmarshal(c, &config)
		if err != nil {
			return &Config{}, err
		}
	}
	config.path = path
	return &config, nil
}
