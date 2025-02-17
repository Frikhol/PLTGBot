package config

type Config struct {
	Name     string `envconfig:"NAME" required:"true"`
	Version  string `envconfig:"VERSION" required:"true"`
	Port     string `envconfig:"GRPC_PORT" default:"50051"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"` // Уровень логирования
	Token    string `envconfig:"TOKEN" required:"true"`     //bot token
}
