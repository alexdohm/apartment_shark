package users

import "apartmenthunter/internal/config"

type UserConfig struct {
	UserID      string
	ZipCodes    []string
	WbsRequired bool
	MinSqm      int
	MaxSqm      int
	MinPrice    int
	MaxPrice    int
}

type FilterConfig struct {
	Users []UserConfig
}

func LoadFromStaticConfig() *FilterConfig {
	return &FilterConfig{Users: []UserConfig{
		{
			UserID:      "",
			ZipCodes:    config.ZipCodes,
			WbsRequired: config.Wbs,
			MinSqm:      config.MinSqm,
			MaxSqm:      config.MaxSqm,
			MinPrice:    config.MinWarm,
			MaxPrice:    config.MaxWarm,
		},
	}}
}
