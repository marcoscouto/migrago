package config

type Config struct {
	MigrationPattern string
}

func DefaultConfig() *Config {
	return &Config{
		MigrationPattern: `^(\d+)_([a-zA-Z0-9_-]+)\.sql$`,
	}
}
