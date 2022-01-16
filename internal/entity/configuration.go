package entity

// Configuration contains the configuration, taken from a
// YAML configuration file
type Configuration struct {
	Database struct {
		Driver string `yaml:"driver" envconfig:"db_driver"`
		DSN    string `yaml:"dsn" envconfig:"db_dsn"`
	} `yaml:"database"`
	Tls struct {
		CertFile string `yaml:"certfile" envconfig:"tls_certfile"`
		KeyFile  string `yaml:"keyfile" envconfig:"tls_keyfile"`
	}
}
