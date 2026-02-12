package core

type Config struct {
	Minimum int `toml:"minimum"`
}

func DefaultConfig() Config {
	return Config{
		Minimum: 300,
	}
}
