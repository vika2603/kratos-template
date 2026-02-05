package log

type Config struct {
	Level       string          `json:"level" yaml:"level"`
	Format      string          `json:"format" yaml:"format"`
	Development bool            `json:"development" yaml:"development"`
	Caller      bool            `json:"caller" yaml:"caller"`
	Sampling    *SamplingConfig `json:"sampling,omitempty" yaml:"sampling,omitempty"`
	Async       *AsyncConfig    `json:"async,omitempty" yaml:"async,omitempty"`
}

type SamplingConfig struct {
	Enabled    bool `json:"enabled" yaml:"enabled"`
	Initial    int  `json:"initial" yaml:"initial"`
	Thereafter int  `json:"thereafter" yaml:"thereafter"`
}

type AsyncConfig struct {
	Enabled       bool `json:"enabled" yaml:"enabled"`
	BufferSize    int  `json:"buffer_size" yaml:"buffer_size"`
	FlushInterval int  `json:"flush_interval" yaml:"flush_interval"`
}

func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: "json",
		Caller: true,
		Async:  DefaultAsyncConfig(),
	}
}

func DevelopmentConfig() Config {
	return Config{
		Level:       "debug",
		Format:      "console",
		Development: true,
		Caller:      true,
		Async:       DefaultAsyncConfig(),
	}
}

func DefaultAsyncConfig() *AsyncConfig {
	return &AsyncConfig{
		Enabled:       true,
		BufferSize:    256 * 1024,
		FlushInterval: 30000,
	}
}

func (c *Config) applyDefaults() {
	if c.Level == "" {
		c.Level = "info"
	}
	if c.Format == "" {
		c.Format = "json"
	}
	if c.Async == nil {
		c.Async = DefaultAsyncConfig()
		return
	}
	if c.Async.BufferSize == 0 {
		c.Async.BufferSize = 256 * 1024
	}
	if c.Async.FlushInterval == 0 {
		c.Async.FlushInterval = 30000
	}
}
