package log

type Config struct {
	Level       string          `json:"level" yaml:"level"`
	Format      string          `json:"format" yaml:"format"`
	Development bool            `json:"development" yaml:"development"`
	Caller      bool            `json:"caller" yaml:"caller"`
	Sampling    *SamplingConfig `json:"sampling,omitempty" yaml:"sampling,omitempty"`
}

type SamplingConfig struct {
	Enabled    bool `json:"enabled" yaml:"enabled"`
	Initial    int  `json:"initial" yaml:"initial"`
	Thereafter int  `json:"thereafter" yaml:"thereafter"`
}

func defaultConfig() Config {
	return Config{
		Level:  "info",
		Format: "json",
		Caller: true,
	}
}

func (c *Config) applyDefaults() {
	if c.Level == "" {
		c.Level = "info"
	}
	if c.Format == "" {
		c.Format = "json"
	}
}
