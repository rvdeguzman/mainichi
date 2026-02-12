package adapters

import (
	"encoding/json"
	"mainichi/core"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Store struct {
	BasePath string
}

func NewStore(basePath string) Store {
	return Store{BasePath: basePath}
}

func DefaultStore() (Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Store{}, err
	}
	return Store{BasePath: filepath.Join(home, ".mainichi")}, nil
}

func (s Store) EnsureDirs() error {
	return os.MkdirAll(filepath.Join(s.BasePath, "entries"), 0o755)
}

func (s Store) entryPath(date string) string {
	return filepath.Join(s.BasePath, "entries", date+".md")
}

func (s Store) LoadEntry(date string) (core.Entry, error) {
	data, err := os.ReadFile(s.entryPath(date))
	if err != nil {
		return core.Entry{}, err
	}
	return core.ParseEntry(data)
}

func (s Store) SaveEntry(e core.Entry) error {
	if err := s.EnsureDirs(); err != nil {
		return err
	}
	data := core.SerializeEntry(e)
	return os.WriteFile(s.entryPath(e.Date), data, 0o644)
}

func (s Store) LoadConfig() (core.Config, error) {
	cfg := core.DefaultConfig()
	path := filepath.Join(s.BasePath, "config.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (s Store) SaveConfig(cfg core.Config) error {
	path := filepath.Join(s.BasePath, "config.toml")
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s Store) EntryExists(date string) bool {
	_, err := os.Stat(s.entryPath(date))
	return err == nil
}

func (s Store) LoadPrompts() ([]string, error) {
	path := filepath.Join(s.BasePath, "prompts.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	return core.LoadPrompts(lines), nil
}

func (s Store) LoadDeckState() (core.Deck, error) {
	path := filepath.Join(s.BasePath, "prompt_state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return core.Deck{}, err
	}
	var deck core.Deck
	if err := json.Unmarshal(data, &deck); err != nil {
		return core.Deck{}, err
	}
	return deck, nil
}

func (s Store) SaveDeckState(deck core.Deck) error {
	path := filepath.Join(s.BasePath, "prompt_state.json")
	data, err := json.MarshalIndent(deck, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s Store) EnsureDefaultPrompts(defaultPrompts string) error {
	path := filepath.Join(s.BasePath, "prompts.txt")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(defaultPrompts), 0o644)
}
