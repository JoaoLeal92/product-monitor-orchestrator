package config

// Config app config
type Config struct {
	Db       DBConfig      `mapstructure:"db"`
	Crawlers CrawlerConfig `mapstructure:"crawlers"`
}

// DBConfig database configs
type DBConfig struct {
	User      string `mapstructure:"user"`
	Name      string `mapstructure:"name"`
	Password  string `mapstructure:"password"`
	Port      int    `mapstructure:"port"`
	Host      string `mapstructure:"host"`
	SilentLog bool   `mapstructure:"silent-log-mode"`
}

type CrawlerConfig struct {
	Amazon string `mapstructure:"amazon"`
}
