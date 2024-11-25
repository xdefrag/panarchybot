package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	OpenAI struct {
		Model    string `toml:"model"`
		Question string `toml:"question"`
	} `toml:"open_ai"`
	Telegram struct {
		FollowUp struct {
			Message string `toml:"message"`
			URL     string `toml:"url"`
		} `toml:"follow_up"`
		Suggest struct {
			SuggestMessage   string `toml:"suggest_message"`
			SuggestedMessage string `toml:"suggested_message"`
		} `toml:"suggest"`
		MainChannelID int64 `toml:"main_channel_id"`
		SuggestChatID int64 `toml:"suggest_chat_id"`
	} `toml:"telegram"`
}

func Get() (*Config, error) {
	cfg := &Config{}

	cd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	paths := []string{
		fmt.Sprintf("%s/panarchybot.dev.toml", cd),
		fmt.Sprintf("%s/panarchybot.toml", cd),
		fmt.Sprintf("%s/panarchybot/panarchybot.toml", os.Getenv("XDG_CONFIG_HOME")),
		"/etc/panarchybot/panarchybot.toml",
	}

	for _, path := range paths {
		file, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}

		if err := toml.Unmarshal(file, cfg); err != nil {
			return nil, err
		}

		return cfg, nil
	}

	return cfg, nil
}
