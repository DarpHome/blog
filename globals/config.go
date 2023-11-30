package globals

type AppConfig struct {
	SecretKey  string `json:"secret_key"`
	AppName    string `json:"app_name"`
	PgHost     string `json:"pghost"`
	PgPort     int    `json:"pgport"`
	PgUser     string `json:"pguser"`
	PgPassword string `json:"pgpassword"`
	PgDbname   string `json:"pgdbname"`
}

var Config AppConfig
