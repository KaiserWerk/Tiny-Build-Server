package global

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

var (
	Version     = "0.0.0-dev"
	VersionDate = "0000-00-00 00:00:00 +00:00"
)

func GetTestConfiguration() *entity.Configuration {
	return &entity.Configuration{
		Database: struct {
			Driver string `yaml:"driver" envconfig:"db_driver"`
			DSN    string `yaml:"dsn" envconfig:"db_dsn"`
		}{
			Driver: "mysql",
			DSN:    "root:root@tcp(127.0.0.1:3306/tinybuildserver",
		},
		Tls: struct {
			CertFile string `yaml:"certfile" envconfig:"tls_certfile"`
			KeyFile  string `yaml:"keyfile" envconfig:"tls_keyfile"`
		}{},
	}
}
