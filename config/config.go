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
	Metrics struct {
		Addr string `toml:"addr"`
	} `toml:"metrics"`
	Stellar struct {
		FundAccount struct {
			Airdrop struct {
				ByUsernameAmount map[string]string `toml:"by_username_amount"`
				IDLessThanAmount map[string]string `toml:"id_less_than_amount"`
				Enable           bool              `toml:"enable"`
			} `toml:"airdrop"`
			Address        string `toml:"address"`
			Seed           string `toml:"seed"`
			DefaultAmount  string `toml:"default_amount"`
			DefaultAirdrop string `toml:"default_airdrop"`
			Network        string `toml:"network"`
			Passphrase     string `toml:"passphrase"`
			Memo           string `toml:"memo"`
			AssetCode      string `toml:"asset_code"`
			AssetIssuer    string `toml:"asset_issuer"`
			BaseFee        int64  `toml:"base_fee"`
		} `toml:"fund_account"`
	} `toml:"stellar"`
	Telegram struct {
		Welcome struct {
			ChannelURL  string `toml:"channel_url"`
			ChannelName string `toml:"channel_name"`
		} `toml:"welcome"`
		MainChannelID int64 `toml:"main_channel_id"`
		SuggestChatID int64 `toml:"suggest_chat_id"`
		Private       struct {
			Enable bool `toml:"enable"`
		} `toml:"private"`
		Thanks struct {
			Enable bool `toml:"enable"`
		} `toml:"thanks"`
	} `toml:"telegram"`
}

func Get(paths ...string) (*Config, error) {
	cfg := &Config{}

	cd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	defaultPaths := []string{
		fmt.Sprintf("%s/panarchybot.dev.toml", cd),
		fmt.Sprintf("%s/panarchybot.toml", cd),
		fmt.Sprintf("%s/panarchybot/panarchybot.toml", os.Getenv("XDG_CONFIG_HOME")),
		"/etc/panarchybot/panarchybot.toml",
	}

	paths = append(paths, defaultPaths...)

	for _, path := range paths {
		if path == "" {
			continue
		}

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
