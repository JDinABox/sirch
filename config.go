package sirch

import "errors"

type Config struct {
	Addr        string
	OpenAIKey   string
	SearxngHost string
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

func NewConfig(options ...Option) (*Config, error) {
	conf := &Config{
		Addr: ":8080",
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
