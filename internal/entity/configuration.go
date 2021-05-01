package entity

// Configuration contains the configuration, taken from a
// YAML configuration file
type Configuration struct {
	Database struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`
	Tls struct {
		Enabled  bool   `yaml:"enabled"`
		CertFile string `yaml:"certfile"`
		KeyFile  string `yaml:"keyfile"`
	}
}
