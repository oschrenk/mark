package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Target is one Prometheus that accepts remote write.
type Target struct {
	Default bool   `toml:"default"`
	URL     string `toml:"url"`
	Metric  string `toml:"metric"`
}

type Config struct {
	Targets map[string]Target `toml:"targets"`
}

// DefaultMetric is used when a target does not name one.
const DefaultMetric = "homelab_event"

func Load() (*Config, error) {
	var c Config
	p := Path()

	if _, err := toml.DecodeFile(p, &c); err != nil {
		return nil, fmt.Errorf("loading config %s: %w", p, err)
	}

	if len(c.Targets) == 0 {
		return nil, fmt.Errorf("no targets configured in %s", p)
	}

	defaults := 0
	for name, t := range c.Targets {
		if t.URL == "" {
			return nil, fmt.Errorf("target %q: url is required", name)
		}
		t.URL = os.ExpandEnv(t.URL)
		if t.Metric == "" {
			t.Metric = DefaultMetric
		}
		c.Targets[name] = t
		if t.Default {
			defaults++
		}
	}
	if defaults > 1 {
		return nil, fmt.Errorf("multiple targets marked as default, expected at most one")
	}

	return &c, nil
}

func (c *Config) Target(name string) (*Target, error) {
	if name == "" {
		for _, t := range c.Targets {
			if t.Default {
				return &t, nil
			}
		}
		if len(c.Targets) == 1 {
			for _, t := range c.Targets {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("no default target configured")
	}
	t, ok := c.Targets[name]
	if !ok {
		return nil, fmt.Errorf("target %q not found in config", name)
	}
	return &t, nil
}

// Path is exported so an error can name the file the user has to create.
func Path() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mark", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mark", "config.toml")
}
