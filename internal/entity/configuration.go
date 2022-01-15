package entity

// Configuration contains the configuration, taken from a
// YAML configuration file
type Configuration struct {
	Database struct {
		Driver string `yaml:"driver" envconfig:"DB_DRIVER"`
		DSN    string `yaml:"dsn" envconfig:"DB_DSNR"`
	} `yaml:"database"`
	Tls struct {
		CertFile string `yaml:"certfile" envconfig:"TLS_CERTFILE"`
		KeyFile  string `yaml:"keyfile" envconfig:"TLS_KEYFILE"`
	}
}
