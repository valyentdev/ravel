package config

type ServerConfig struct {
	CorrosionServer string `json:"corrosion_server"`
	PGDatabaseUrl   string `json:"pg_database_url"`
}
