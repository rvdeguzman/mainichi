package core

type Config struct {
	Minimum      int    `toml:"minimum"`
	AutoPrompt   bool   `toml:"auto_prompt"`
	PromptSource string `toml:"prompt_source"`
}

func DefaultConfig() Config {
	return Config{
		Minimum:      250,
		PromptSource: "deck",
	}
}
