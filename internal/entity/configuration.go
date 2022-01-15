package entity

// Configuration contains the configuration, taken from a
// YAML configuration file
type Configuration struct {
	Database struct {
		Driver string `yaml:"driver" envconfig:"TBS_DB_DRIVER"`
		DSN    string `yaml:"dsn" envconfig:"TBS_DB_DSNR"`
	} `yaml:"database"`
	Tls struct {
		CertFile string `yaml:"certfile" envconfig:"TBS_TLS_CERTFILE"`
		KeyFile  string `yaml:"keyfile" envconfig:"TBS_TLS_KEYFILE"`
	}
}
