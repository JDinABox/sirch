package sirch

import (
	"errors"
	"fmt"
	"strconv"
)

type Config struct {
	Addr        string
	OpenAIKey   string
	SearxngHost string
	Public      bool
	DBPath      string
}

type Option func(*Config) error

func WithAddr(addr string) Option {
	return func(c *Config) error {
		c.Addr = addr
		return nil
	}
}

func WithOpenAIKey(key string) Option {
	return func(c *Config) error {
		c.OpenAIKey = key
		return nil
	}
}
func WithSearxngHost(host string) Option {
	return func(c *Config) error {
		c.SearxngHost = host
		return nil
	}
}

func WithPublicString(public string) Option {
	return func(c *Config) error {
		boolValue, err := strconv.ParseBool(public)
		if err != nil {
			return fmt.Errorf("unable to parse public environment variable: %w", err)
		}
		c.Public = boolValue
		return nil
	}
}

func WithDBPath(path string) Option {
	return func(c *Config) error {
		c.DBPath = path
		return nil
	}
}

func NewConfig(options ...Option) (*Config, error) {
	conf := &Config{
		Addr:   ":8080",
		Public: true,
		DBPath: "./db/db.sqlite",
	}
	for _, o := range options {
		if err := o(conf); err != nil {
			return nil, err
		}
	}

	if conf.OpenAIKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}
	if conf.SearxngHost == "" {
		return nil, errors.New("SEARXNG_HOST is not set")
	}

	return conf, nil
}
