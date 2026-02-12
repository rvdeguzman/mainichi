package core

type Config struct {
	Minimum    int  `toml:"minimum"`
	AutoPrompt bool `toml:"auto_prompt"`
}

func DefaultConfig() Config {
	return Config{
		Minimum: 300,
	}
}
