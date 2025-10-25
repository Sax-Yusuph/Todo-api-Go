package application

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sst/sst/v3/sdk/golang/resource"
)

type Config struct {
	JWTSecret    string
	RedisOptions *redis.Options
}

func LoadConfig() (Config, error) {
	secrets := []string{"JWTSecret", "RedisUrl"}

	config := Config{}

	for _, secret := range secrets {
		value, err := resource.Get(secret, "value")
		if err != nil {
			return config, fmt.Errorf("Could not load %s: %w", secret, err)

		}

		switch secret {
		case "JWTSecret":
			config.JWTSecret = value.(string)
		case "RedisUrl":
			redisOptions, err := redis.ParseURL(value.(string))
			if err != nil {
				return config, fmt.Errorf("Could not parse Redis URL: %w", err)
			}

			config.RedisOptions = redisOptions
		default:
			return config, fmt.Errorf("Unknown secret: %s", secret)
		}
	}

	return config, nil

}
