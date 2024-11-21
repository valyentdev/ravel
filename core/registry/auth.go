package registry

type RegistryAuthConfig struct {
	ServerAddress string `json:"server_address"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Auth          string `json:"auth"`
	IdentityToken string `json:"identitytoken"`
}
