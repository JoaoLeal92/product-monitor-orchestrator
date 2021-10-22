package config

// Config app config
type Config struct {
	Db       DBConfig      `mapstructure:"db"`
	Crawlers CrawlerConfig `mapstructure:"crawlers"`
	Log      LogConfig     `mapstructure:"log"`
	Queue    QueueConfig   `mapstructure:"queue"`
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
	Amazon       string `mapstructure:"amazon"`
	MercadoLivre string `mapstructure:"mercado-livre"`
	NumCrawlers  int    `mapstructure:"num-crawlers"`
}

type LogConfig struct {
	Host       string `mapstructure:"host"`
	Port       string `mapstructure:"port"`
	Db         string `mapstructure:"db"`
	Collection string `mapstructure:"collection"`
}

type QueueConfig struct {
	QueueName string `mapstructure:"queue-name"`
}
