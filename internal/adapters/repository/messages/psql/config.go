package psql

type PostgresMessageConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT" env-required:"true"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-required:"true"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `yaml:"db_name" env:"POSTGRES_DB_NAME" env-required:"true"`
}
