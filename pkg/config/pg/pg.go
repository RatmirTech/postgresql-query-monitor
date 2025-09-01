package pg

type DbConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode  string
}
