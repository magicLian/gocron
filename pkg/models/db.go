package models

type DatabaseConfig struct {
	Type             string
	ConnectionString string
	User             string
	Password         string
	Host             string
	Name             string
	SchemaName       string
	GroupName        string
	MaxOpenConn      int
	MaxIdleConn      int
	ConnMaxLifetime  int
	LogQueries       bool
	UrlQueryParams   map[string][]string
	SslMode          string
	CaCertPath       string
	ClientKeyPath    string
	ClientCertPath   string

	//sqlite3
	Path      string
	CacheMode string
}
